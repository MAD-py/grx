package grx

import (
	"log"
	"time"

	"github.com/MAD-py/grx/pkg/config"
)

type grx struct {
	servers []*server
}

func (g *grx) Stop() {
	log.Printf("Stopping grx...")
	for _, s := range g.servers {
		if s.status == online {
			go s.shutdown()
		}
	}

Loop:
	for {
		time.Sleep(time.Second * 2)
		for _, s := range g.servers {
			if s.status != offline {
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

func New(grxConfig *config.Config) (*grx, error) {
	servers := make([]*server, len(grxConfig.Servers))
	for i, srv := range grxConfig.Servers {
		server, err := newServer(srv)
		if err != nil {
			return nil, err
		}
		servers[i] = server
	}
	return &grx{servers: servers}, nil
}
