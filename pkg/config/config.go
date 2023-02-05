package config

import "time"

type Server struct {
	ProxyID           string
	LogName           string
	Listen            string
	Pattern           string
	MaxConnections    int
	TimeoutPerRequest time.Duration
}
