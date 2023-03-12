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

	TimeoutPerRequest time.Duration
}

type StaticServer struct {
	Server

	PathPrefix string
}

type Config struct {
	Servers []interface{}
}
