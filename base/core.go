package base

import (
	"context"

	"github.com/Pororochenzy/lego/core"
	"github.com/Pororochenzy/lego/sys/rpc"
)

type Result struct {
	Index  string
	Result interface{}
	Err    error
}

type ISingleService interface {
	core.IService
}

type IClusterService interface {
	core.IService

	GetCategory() core.S_Category //服务类别 例如游戏服
	GetRpcId() string             //获取rpc通信id
	GetPreWeight() float64
	GetIp() string                                                                                                                        //获取ip
	GetTag() string                                                                                                                       //获取集群标签
	Register(rcvr interface{}) (err error)                                                                                                // 注册服务对象
	RegisterFunction(fn interface{}) (err error)                                                                                          // 注册服务方法
	RegisterFunctionName(name string, fn interface{}) (err error)                                                                         //注册服务方法 自定义服务名
	RpcCall(ctx context.Context, servicePath, serviceMethod string, args interface{}, reply interface{}) (err error)                      //调用远端服务接口 同步
	RpcGo(ctx context.Context, servicePath, serviceMethod string, args interface{}, reply interface{}) (call *rpc.MessageCall, err error) //调用远端服务接口 异步
	Broadcast(ctx context.Context, servicePath, serviceMethod string, args interface{}) (err error)                                       //广播调用远端服务接口
	RpcInvokeById(sId string, rkey core.Rpc_Key, iscall bool, arg ...interface{}) (result interface{}, err error)
	Subscribe(id core.Rpc_Key, f interface{}) (err error)
	RpcInvokeByType(sType string, rkey core.Rpc_Key, iscall bool, arg ...interface{}) (result interface{}, err error)

	GetSessionsByCategory(category core.S_Category) (ss []core.IServiceSession) //按服务类别获取服务列表
}
