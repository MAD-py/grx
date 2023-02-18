package grx

import (
	"errors"
	"net"
	"time"

	"github.com/MAD-py/grx/pkg/config"
)

type grx struct {
	servers []*server

	status grxStatus

	listener net.Listener
}

func (g *grx) stop() {
	for _, s := range g.servers {
		if s.status == online {
			go s.shutdown()
		}
	}

	go func() {
	Loop:
		for {
			time.Sleep(time.Second * 5)
			for _, s := range g.servers {
				if s.status != offline {
					continue Loop
				}
			}
			g.status = stopped
		}
	}()
	g.status = stopping
}

func (g *grx) run() error {
	if g.status == running {
		return errors.New("already running")
	}
	if g.status == stopping {
		return errors.New(
			"there are still services that have not yet been fully stopped",
		)
	}

	for _, s := range g.servers {
		go s.run()
	}
	g.status = running
	return nil
}

func (g *grx) listen() error {
	listener, err := net.Listen("tcp", ":48010")
	if err != nil {
		return err
	}
	g.listener = listener

	go func() {
		for {
			conn, err := g.listener.Accept()
			if err != nil {
				panic(err)
			}
			go g.apply(newConn(conn))
		}
	}()
	return nil
}

func (g *grx) apply(conn *connection) {
	defer conn.close()
	message, s := conn.read()
	if !s {
		return
	}

	var m string
	var body map[string]interface{}
	switch message.Action {
	case run:
		err := g.run()
		if err != nil {
			m = err.Error()
		} else {
			m = "OK"
		}
	case stop:
		g.stop()
		m = "OK"
	case kill:
		g.listener.Close()
		g.stop()
		return
	}

	conn.write(m, body)
}

func new(srvs []*config.Server) (*grx, error) {
	servers := make([]*server, len(srvs))
	for i, srv := range srvs {
		server, err := newServer(srv)
		if err != nil {
			return nil, err
		}
		servers[i] = server
	}

	return &grx{
		servers: servers,
		status:  stopped,
	}, nil
}

func StartProxy(srvs []*config.Server) error {
	grx, err := new(srvs)
	if err != nil {
		return err
	}

	err = grx.listen()
	if err != nil {
		return err
	}

	grx.status = awaitting
	grx.run()
	return nil
}
