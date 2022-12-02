package cbase

import (
	"github.com/Pororochenzy/lego/core"
	"github.com/Pororochenzy/lego/sys/registry"
	"github.com/Pororochenzy/lego/sys/rpc"
)

func NewServiceSession(node *registry.ServiceNode) (ss core.IServiceSession, err error) {
	session := new(ServiceSession)
	session.node = node

	optRpc := rpc.SetServiceNode(&core.ServiceNode{ //新建了一个rpc客户端
		Tag:     node.Tag,
		Type:    node.Type,
		Id:      node.Id,
		Version: node.Version,
		Addr:    node.IP,
	})
	session.rpc, err = rpc.NewSys(optRpc)
	ss = session
	return
}

type ServiceSession struct {
	node *registry.ServiceNode
	//rpc  rpc.IRpcClient
	rpc rpc.ISys
}

func (this *ServiceSession) GetId() string {
	return this.node.Id
}
func (this *ServiceSession) GetIp() string {
	return this.node.IP
}
func (this *ServiceSession) GetRpcId() string {
	return this.node.RpcId
}

func (this *ServiceSession) GetType() string {
	return this.node.Type
}
func (this *ServiceSession) GetVersion() string {
	return this.node.Version
}
func (this *ServiceSession) SetVersion(v string) {
	this.node.Version = v
}

func (this *ServiceSession) GetPreWeight() float64 {
	return this.node.PreWeight
}
func (this *ServiceSession) SetPreWeight(p float64) {
	this.node.PreWeight = p
}
func (this *ServiceSession) Done() {
	this.rpc.Close()
}
func (this *ServiceSession) Call(f core.Rpc_Key, params ...interface{}) (interface{}, error) {
	//return this.rpc.Call(string(f), params...)
	//return this.rpc.Call(context.Background(), servicePath string,string(f), args interface{}, reply interface{}) (err error)
	return nil, nil
}
func (this *ServiceSession) CallNR(f core.Rpc_Key, params ...interface{}) (err error) {
	//return this.rpc.CallNR(string(f), params...)
	return nil
}
