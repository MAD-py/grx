package main

import (
	"time"

	"github.com/MAD-py/grx/pkg/config"
	"github.com/MAD-py/grx/pkg/grx"
)

func main() {
	shutdown := make(chan uint8)
	config := config.Server{
		ProxyID:           "192.168.1.0/24",
		LogName:           "Main",
		Listen:            "localhost:8080",
		Pattern:           "localhost:8001",
		MaxConnections:    100,
		TimeoutPerRequest: 40,
	}
	server, err := grx.NewServer(&config, shutdown)
	if err != nil {
		println(err.Error())
		panic(err)
	}

	go server.Run()
	time.Sleep(30 * time.Second)
	shutdown <- 1
	time.Sleep(10 * time.Second)
}
