package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/MAD-py/grx/pkg/config"
	"github.com/MAD-py/grx/pkg/grx"
)

func main() {
	filePath := flag.String("file", "grx.yml", "Configuration file path")
	flag.Parse()

	grxServers, err := config.Load(*filePath)
	if err != nil {
		panic(err)
	}

	proxy, err := grx.New(grxServers)
	if err != nil {
		panic(err)
	}

	proxy.Run()
	sgn := make(chan os.Signal, 1)
	signal.Notify(sgn, syscall.SIGINT)
	<-sgn
	proxy.Stop()
}
