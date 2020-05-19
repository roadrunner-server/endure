package rpc

import "net/rpc"

type RPC struct {
	rpc rpc.Server
}

type RPCService interface {
	Name() string
	RCPService() interface{}
}

func (r *RPC) Init() error {
	return nil
}

func (r *RPC) Registers() []interface{} {
	return []interface{}{r.Register}
}

func (r *RPC) Register(svc RPCService) error {
	return r.rpc.RegisterName(
		svc.Name(),
		svc.RCPService(),
	)
}
