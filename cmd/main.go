package main

import (
	"time"

	"github.com/MAD-py/grx/pkg/config"
	"github.com/MAD-py/grx/pkg/grx"
	"github.com/MAD-py/grx/pkg/notification"
)

func main() {
	config := config.Server{
		ProxyID:           "192.168.1.0/24",
		LogName:           "Main",
		Listen:            "localhost:8080",
		Pattern:           "localhost:8001",
		MaxConnections:    100,
		TimeoutPerRequest: 40,
	}
	notifer, subscriber := notification.NewNotifierAndSubscriber()
	server, err := grx.NewServer(&config, subscriber)
	if err != nil {
		println(err.Error())
		panic(err)
	}

	go server.Run()
	time.Sleep(30 * time.Second)
	notifer.Sender <- notification.Shutdown
	time.Sleep(10 * time.Second)
}
