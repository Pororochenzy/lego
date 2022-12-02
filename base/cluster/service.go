package cluster

import (
	"context"
	"fmt"
	"sync"

	"github.com/Pororochenzy/lego/base"
	"github.com/Pororochenzy/lego/core"
	"github.com/Pororochenzy/lego/core/cbase"
	"github.com/Pororochenzy/lego/sys/cron"
	"github.com/Pororochenzy/lego/sys/event"
	"github.com/Pororochenzy/lego/sys/log"
	"github.com/Pororochenzy/lego/sys/proto"
	"github.com/Pororochenzy/lego/sys/registry"
	"github.com/Pororochenzy/lego/sys/rpc"
	"github.com/Pororochenzy/lego/sys/timewheel"
)

// 集群服务数据
type ClusterService struct {
	cbase.ServiceBase
	option         *Options
	serviceNode    *core.ServiceNode
	clusterService base.IClusterService
	serverList     sync.Map
}

func (this *ClusterService) GetTag() string {
	return this.option.Setting.Tag
}
func (this *ClusterService) GetId() string {
	return this.option.Setting.Id
}
func (this *ClusterService) GetType() string {
	return this.option.Setting.Type
}
func (this *ClusterService) GetVersion() string {
	return this.option.Version
}
func (this *ClusterService) GetSettings() core.ServiceSttings {
	return this.option.Setting
}
func (this *ClusterService) Options() *Options {
	return this.option
}
func (this *ClusterService) Configure(option ...Option) {
	this.option = newOptions(option...)
	this.serviceNode = &core.ServiceNode{
		Tag:     this.option.Setting.Tag,
		Id:      this.option.Setting.Id,
		Type:    this.option.Setting.Type,
		Version: this.option.Version,
		Meta:    make(map[string]string),
	}
}

// 集群服务初始化
func (this *ClusterService) Init(service core.IService) (err error) {
	this.clusterService = service.(base.IClusterService)
	return this.ServiceBase.Init(service)
}

// 初始化日志，事件，RPC,时间轮，proto 等周边系统
// 在主服务初始化的时候被调用
func (this *ClusterService) InitSys() {
	if err := log.OnInit(this.option.Setting.Sys["log"]); err != nil {
		panic(fmt.Sprintf("初始化log系统失败 err:%v", err))
	} else {
		log.Infof("Sys log Init success !")
	}
	if err := event.OnInit(this.option.Setting.Sys["event"]); err != nil {
		log.Panicf(fmt.Sprintf("初始化event系统失败 err:%v", err))
	} else {
		log.Infof("Sys event Init success !")
	}
	//根据配置初始化RPC
	if err := rpc.OnInit(this.option.Setting.Sys["rpc"], rpc.SetServiceNode(this.serviceNode)); err != nil {
		log.Panicf(fmt.Sprintf("初始化rpc系统 err:%v", err))
	} else {
		log.Infof("Sys rpc Init success !")
	}
	if err := timewheel.OnInit(nil); err != nil {
		panic(fmt.Sprintf("初始化timewheel系统失败 %v", err))
	}
	if err := proto.OnInit(map[string]interface{}{"MsgProtoType": 1}); err != nil {
		panic(fmt.Sprintf("初始化proto失败 %v", err))
	}

	event.Register(core.Event_ServiceStartEnd, func() { //阻塞 先注册服务集群 保证其他服务能及时发现
		if err := rpc.Start(); err != nil {
			log.Panicf(fmt.Sprintf("启动RPC失败 err:%v", err))
		}
	})
}

func (this *ClusterService) Destroy() (err error) {
	if err = rpc.Close(); err != nil {
		return
	}
	cron.Close()
	err = this.ServiceBase.Destroy()
	return
}

// 注册服务对象
func (this *ClusterService) Register(rcvr interface{}) (err error) {
	return rpc.Register(rcvr)
}

// 注册服务方法
func (this *ClusterService) RegisterFunction(fn interface{}) (err error) {
	return rpc.RegisterFunction(fn)
}

// 注册服务方法 自定义服务名
func (this *ClusterService) RegisterFunctionName(name string, fn interface{}) (err error) {
	return rpc.RegisterFunctionName(name, fn)
}

// 调用远端服务接口 同步
func (this *ClusterService) RpcCall(ctx context.Context, servicePath, serviceMethod string, args interface{}, reply interface{}) (err error) {
	err = rpc.Call(ctx, servicePath, serviceMethod, args, reply)
	return
}

// 调用远端服务接口 异步
func (this *ClusterService) RpcGo(ctx context.Context, servicePath, serviceMethod string, args interface{}, reply interface{}) (call *rpc.MessageCall, err error) {
	return rpc.Go(ctx, servicePath, serviceMethod, args, reply)
}

// 广播调用远端服务接口
func (this *ClusterService) Broadcast(ctx context.Context, servicePath, serviceMethod string, args interface{}) (err error) {
	return rpc.Broadcast(ctx, servicePath, serviceMethod, args)
}
func (this *ClusterService) RpcInvokeById(sId string, rkey core.Rpc_Key, iscall bool, arg ...interface{}) (result interface{}, err error) {
	ss, ok := this.serverList.Load(sId)
	if !ok {
		if node, err := registry.GetServiceById(sId); err != nil {
			log.Errorf("未找到目标服务【%s】节点 err:%s", sId, err.Error())
			return nil, fmt.Errorf("No Found " + sId)
		} else {
			ss, err = cbase.NewServiceSession(node)
			if err != nil {
				return nil, fmt.Errorf(fmt.Sprintf("创建服务会话失败【%s】 err = %s", sId, err.Error()))
			} else {
				this.serverList.Store(node.Id, ss)
			}
		}
	}
	if iscall {
		result, err = ss.(core.IServiceSession).Call(rkey, arg...)
	} else {
		err = ss.(core.IServiceSession).CallNR(rkey, arg...)
	}
	return
	//defer lego.Recover(fmt.Sprintf("RpcInvokeById sId:%s rkey:%v iscall %v arg %v", sId, rkey, iscall, arg))
	//this.lock.RLock()
	//defer this.lock.RUnlock()
	//ss, ok := this.serverList.Load(sId)
	//if !ok {
	//	if node, err := registry.GetServiceById(sId); err != nil {
	//		log.Errorf("未找到目标服务【%s】节点 err:%v", sId, err)
	//		return nil, fmt.Errorf("No Found " + sId)
	//	} else {
	//		ss, err = cbase.NewServiceSession(node)
	//		if err != nil {
	//			return nil, fmt.Errorf(fmt.Sprintf("创建服务会话失败【%s】 err:%v", sId, err))
	//		} else {
	//			this.serverList.Store(node.Id, ss)
	//		}
	//	}
	//}
	//if iscall {
	//	result, err = ss.(core.IServiceSession).Call(rkey, arg...)
	//} else {
	//	err = ss.(core.IServiceSession).CallNR(rkey, arg...)
	//}
	//return "", nil
}

func (this *ClusterService) Subscribe(id core.Rpc_Key, f interface{}) (err error) {
	//rpc.Go(id, f)
	//err = registry.PushServiceInfo()
	return nil
}
func (this *ClusterService) RpcInvokeByType(sType string, rkey core.Rpc_Key, iscall bool, arg ...interface{}) (result interface{}, err error) {
	//defer lego.Recover(fmt.Sprintf("RpcInvokeByType sType:%s rkey:%v iscall %v arg %v", sType, rkey, iscall, arg))
	//this.lock.RLock()
	//defer this.lock.RUnlock()
	//ss, err := this.c.DefauleRpcRouteRules(sType, core.AutoIp)
	//if err != nil {
	//	log.Errorf("未找到目标服务【%s】节点 err:%v", sType, err)
	//	return nil, err
	//}
	//if iscall {
	//	result, err = ss.Call(rkey, arg...)
	//} else {
	//	err = ss.CallNR(rkey, arg...)
	//}
	return nil, nil
}
func (this *ClusterService) GetCategory() core.S_Category { //服务类别 例如游戏服
	return ""
}

func (this *ClusterService) GetRpcId() string { //获取rpc通信id
	return ""
}
func (this *ClusterService) GetPreWeight() float64 {
	return 0
}
func (this *ClusterService) GetIp() string { //获取ip
	return ""
}
func (this *ClusterService) GetSessionsByCategory(category core.S_Category) (ss []core.IServiceSession) {
	ss = make([]core.IServiceSession, 0)
	if nodes := registry.GetServiceByCategory(category); nodes == nil {
		log.Errorf("获取目标类型【%s】服务集失败", category)
		return ss
	} else {
		for _, v := range nodes {
			if s, ok := this.serverList.Load(v.Id); ok {
				ss = append(ss, s.(core.IServiceSession))
			} else {
				s, err := cbase.NewServiceSession(v)
				if err != nil {
					log.Errorf("创建服务会话失败【%s】 err = %s", v.Id, err.Error())
					continue
				} else {
					this.serverList.Store(v.Id, s)
					ss = append(ss, s.(core.IServiceSession))
				}
			}
		}
	}
	return
}
