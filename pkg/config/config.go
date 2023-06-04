package config

import "time"

type Server struct {
	Name           string
	ListenAddr     string
	MaxConnections int
}

type ForwardServer struct {
	Server

	ID string

	LoadBalancer LoadBalancer

	Forward []*Forward

	UseForwarded bool

	TimeoutPerRequest time.Duration
}

type StaticServer struct {
	Server

	PathPrefix string
}

type Forward struct {
	Addr string

	Weight uint8
}

type Servers []any

type LoadBalancer uint8

const (
	non LoadBalancer = iota
	Base
	RoundRobin
	WeightedRoundRobin
)

func (s LoadBalancer) String() string {
	switch s {
	case Base:
		return "Base"
	case RoundRobin:
		return "Round Robin"
	case WeightedRoundRobin:
		return "Weighted Round Robin"
	}
	return "unknown"
}
