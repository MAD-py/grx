package lb

import "github.com/MAD-py/grx/pkg/config"

type LoadBalancer interface {
	GetServer() string
}

type Base struct {
	server *config.Forward
}

func (a *Base) GetServer() string { return a.server.Addr }

func NewBase(server *config.Forward) *Base { return &Base{server: server} }
