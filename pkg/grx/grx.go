package grx

import (
	"errors"
	"log"
	"time"

	"github.com/MAD-py/grx/pkg/config"
)

type grx struct {
	servers []server
}

func (g *grx) Stop() {
	log.Printf("Stopping grx...")
	for _, s := range g.servers {
		if s.getStatus() == online {
			go s.shutdown()
		}
	}

Loop:
	for {
		time.Sleep(time.Second * 2)
		for _, s := range g.servers {
			if s.getStatus() != offline {
				continue Loop
			}
		}
		break Loop
	}
	log.Printf("grx stopped")
}

func (g *grx) Run() {
	log.Printf("Starting grx...")
	for _, s := range g.servers {
		go s.run()
	}
}

func New(grxServers config.Servers) (*grx, error) {
	servers := make([]server, len(grxServers))
	for i, srv := range grxServers {
		switch v := srv.(type) {
		case *config.ForwardServer:
			server, err := newForwardServer(v)
			if err != nil {
				return nil, err
			}
			servers[i] = server
		case *config.StaticServer:
			server, err := newStaticServer(v)
			if err != nil {
				return nil, err
			}
			servers[i] = server
		default:
			return nil, errors.New("unknown server type")
		}
	}
	return &grx{servers: servers}, nil
}
