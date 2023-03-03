package config

import "time"

type Server struct {
	ID   string
	Name string

	ListenAddr  string
	PatternAddr string

	MaxConnections    int
	TimeoutPerRequest time.Duration
}

type Config struct {
	Servers []*Server
}
