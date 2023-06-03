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

	PatternAddr string

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

type Config struct {
	Servers []interface{}
}
