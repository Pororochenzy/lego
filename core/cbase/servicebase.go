package cbase

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"

	"github.com/Pororochenzy/lego"
	"github.com/Pororochenzy/lego/core"
	"github.com/Pororochenzy/lego/sys/event"
	"github.com/Pororochenzy/lego/sys/log"
)

type defaultModule struct {
	seetring map[string]interface{}
	mi       core.IModule
	closeSig chan bool
	wg       sync.WaitGroup
}

func (this *defaultModule) run() {
	this.mi.Run(this.closeSig)
	this.wg.Done()
}
func (this *defaultModule) destroy() (err error) {
	defer lego.Recover(fmt.Sprintf("Module :%s destroy", this.mi.GetType()))
	err = this.mi.Destroy()
	if err != nil {
		err = fmt.Errorf("关闭模块【%s】失败 err:%s", this.mi.GetType(), err.Error())
	}
	return
}

// 主服务数据
type ServiceBase struct {
	closesig chan string
	Service  core.IService
	comps    map[core.S_Comps]core.IServiceComp
	modules  map[core.M_Modules]*defaultModule
}

// 主服务数据初始化
func (this *ServiceBase) Init(service core.IService) (err error) {
	this.closesig = make(chan string, 1)
	this.Service = service
	this.modules = make(map[core.M_Modules]*defaultModule)
	this.Service.InitSys()
	for _, v := range this.comps {
		options := v.NewOptions()
		if o, ok := service.GetSettings().Comps[string(v.GetName())]; ok {
			options.LoadConfig(o)
		}
		err = v.Init(this.Service, v, options)
		if err != nil {
			return
		}
	}
	log.Infof("服务[%s] 初始化完成!", this.Service.GetId())
	return nil
}

// 配置服务组件
func (this *ServiceBase) OnInstallComp(cops ...core.IServiceComp) {
	this.comps = make(map[core.S_Comps]core.IServiceComp)
	for _, v := range cops {
		if _, ok := this.comps[v.GetName()]; ok {
			log.Errorf("覆盖注册组件【%s】", v.GetName())
		}
		this.comps[v.GetName()] = v
	}
}

// 主服务 启动所有组件
func (this *ServiceBase) Start() (err error) {
	for _, v := range this.comps {
		err = v.Start()
		if err != nil {
			return
		}
	}
	log.Infof("服务[%s:%s] 启动完成!", this.Service.GetId(), this.Service.GetVersion())
	return
}

// 主服务实现了运行接口
// 入参为模块
// 目前只有一个网关模块
func (this *ServiceBase) Run(mod ...core.IModule) {
	go func() {
		//根据配置生成组件实例数组
		for _, v := range mod {

			// 获取配置文件中，跟组件相关的数据配置，配置头为组件定义的类别名称，如下
			// modules:
			// SM_GateModule:
			// TcpAddr: "127.0.0.1:3563"
			// WSAddr: "127.0.0.1:3653"
			// HeartbeatInterval: 4
			// MaxHeartStopNum: 3
			if sf, ok := this.Service.GetSettings().Modules[string(v.GetType())]; ok {
				this.modules[v.GetType()] = &defaultModule{
					seetring: sf,                 //配置参数存到了这个MAP里
					mi:       v,                  //模块实例存到了这里
					closeSig: make(chan bool, 1), //关闭信号
				}
				log.Warnf("注册模块【%s】根据配置，初始化完成。", v.GetType())
			} else {
				this.modules[v.GetType()] = &defaultModule{
					seetring: make(map[string]interface{}),
					mi:       v,
					closeSig: make(chan bool, 1),
				}
				log.Warnf("注册模块【%s】 没有对应的配置信息", v.GetType())
			}
		}
		for _, v := range this.modules {
			// 声明个配置对象
			options := v.mi.NewOptions()
			if err := options.LoadConfig(v.seetring); err == nil {
				log.Warnf("加载组件配置信息-> 【%s】", v.seetring)
				err = v.mi.Init(this.Service, v.mi, options)
				if err != nil {
					log.Panicf(fmt.Sprintf("初始化模块【%s】错误 err:%v", v.mi.GetType(), err))
				}
				log.Warnf("初始化组件配置成功")

			} else {
				log.Panicf(fmt.Sprintf("模块【%s】 Options:%v 配置错误 err:%v", v.mi.GetType(), v.seetring, err))
			}
		}
		for _, v := range this.modules {
			err := v.mi.Start()
			if err != nil {
				log.Panicf(fmt.Sprintf("启动模块【%s】错误 err:%v", v.mi.GetType(), err))
			}
		}

		//运行组件
		for _, v := range this.modules {
			v.wg.Add(1)
			go v.run()
		}
		//RPC到服务发现地址,注册自己
		event.TriggerEvent(core.Event_ServiceStartEnd) //广播事件
	}()
	//监听外部关闭服务信号
	c := make(chan os.Signal, 1)
	//添加进程结束信号
	signal.Notify(c,
		os.Interrupt,    //退出信号 ctrl+c退出
		os.Kill,         //kill 信号
		syscall.SIGHUP,  //终端控制进程结束(终端连接断开)
		syscall.SIGINT,  //用户发送INTR字符(Ctrl+C)触发
		syscall.SIGTERM, //结束程序(可以被捕获、阻塞或忽略)
		syscall.SIGQUIT) //用户发送QUIT字符(Ctrl+/)触发
	select {
	case sig := <-c:
		log.Errorf("服务[%s] 关闭 signal = %v\n", this.Service.GetId(), sig)
	case <-this.closesig:
		log.Errorf("服务[%s] 关闭\n", this.Service.GetId())
	}
}

func (this *ServiceBase) Close(closemsg string) {
	this.closesig <- closemsg
}

func (this *ServiceBase) Destroy() (err error) {
	for _, v := range this.modules {
		v.closeSig <- true
		v.wg.Wait()
		err = v.destroy()
		if err != nil {
			return
		}
	}
	for _, v := range this.comps {
		err = v.Destroy()
		if err != nil {
			return
		}
	}
	return
}

func (this *ServiceBase) GetModule(ModuleName core.M_Modules) (module core.IModule, err error) {
	if v, ok := this.modules[ModuleName]; ok {
		return v.mi, nil
	} else {
		return nil, fmt.Errorf("未装配模块【%s】", ModuleName)
	}
}

func (this *ServiceBase) GetComp(CompName core.S_Comps) (comp core.IServiceComp, err error) {
	if v, ok := this.comps[CompName]; ok {
		return v, nil
	} else {
		return nil, fmt.Errorf("Service 未装配组件【%s】", CompName)
	}
}
func Recover() {
	if r := recover(); r != nil {
		buf := make([]byte, 1024)
		l := runtime.Stack(buf, false)
		log.Panicf("%v: %s", r, buf[:l])
	}
}
