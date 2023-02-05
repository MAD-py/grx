package main

import (
	"github.com/MAD-py/grx/pkg/config"
	"github.com/MAD-py/grx/pkg/grx"
)

func main() {
	config := config.Server{
		ProxyID:           "192.168.1.0/24",
		LogName:           "Main",
		Listen:            "localhost:8080",
		Pattern:           "localhost:8001",
		MaxConnections:    100,
		TimeoutPerRequest: 30,
	}
	server, err := grx.NewServer(&config)
	if err != nil {
		println(err.Error())
		panic(err)
	}

	server.Run()
}
