package lb

import "github.com/MAD-py/grx/pkg/config"

type RoundRobin struct {
	servers []*config.Forward

	index uint8

	maxIndex uint8
}

func (lb *RoundRobin) GetServer() string {
	server := lb.servers[lb.index]
	lb.index++
	if lb.index > lb.maxIndex {
		lb.index = 0
	}
	return server.Addr
}

func NewRoundRobin(servers []*config.Forward) *RoundRobin {
	return &RoundRobin{
		servers:  servers,
		index:    0,
		maxIndex: uint8(len(servers) - 1),
	}
}
