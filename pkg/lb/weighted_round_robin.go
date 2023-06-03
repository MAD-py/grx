package lb

import (
	"sort"

	"github.com/MAD-py/grx/pkg/config"
)

type WeightedRoundRobin struct {
	servers []*config.Forward

	index uint8

	maxIndex uint8

	currentWeight uint8
}

func (lb *WeightedRoundRobin) GetServer() string {
	server := lb.servers[lb.index]
	lb.currentWeight--

	if lb.index == lb.maxIndex && lb.currentWeight == 0 {
		lb.index = 0
		lb.currentWeight = lb.servers[0].Weight
	} else if lb.currentWeight == 0 {
		lb.index++
	}
	return server.Addr
}

func NewWeightedRoundRobin(servers []*config.Forward) *WeightedRoundRobin {
	sort.Slice(servers, func(i, j int) bool {
		return servers[i].Weight > servers[j].Weight
	})
	return &WeightedRoundRobin{
		servers:       servers,
		index:         0,
		maxIndex:      uint8(len(servers) - 1),
		currentWeight: servers[0].Weight,
	}
}
