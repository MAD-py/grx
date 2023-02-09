package grx

import (
	"bufio"
	"bytes"
	"log"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/MAD-py/grx/pkg/config"
	"github.com/MAD-py/grx/pkg/errors"
	proxyHTTP "github.com/MAD-py/grx/pkg/http"
	"github.com/MAD-py/grx/pkg/notification"
)

const maxRequestSize = 32 << 20 // INFO: 32 MB

type Server struct {
	// Configuration for the server.
	config *config.Server

	// Channel through which the master notifies the shutdown.
	subscriber *notification.Subscriber

	// HTTP client in charge of processing incoming requests .
	client *http.Client

	// TCP listener to accept incoming connections.
	listener *net.TCPListener

	// Forwarding TCP address
	pattern *net.TCPAddr

	// Connections are limited and this channel is used as a Semaphore
	// to prevent overloading.
	connections chan uint8
}

func (s *Server) forward(conn *net.TCPConn) {
	defer func() {
		conn.Close()
		log.Printf(
			"%s => Close connection [%s]",
			s.config.LogName, conn.RemoteAddr().String(),
		)
		<-s.connections
	}()

	req, err := http.ReadRequest(bufio.NewReaderSize(conn, maxRequestSize))
	if err != nil {
		res := proxyHTTP.ErrorToResponse(nil, errors.BadRequest())
		b := bytes.Buffer{}
		res.IntoForwarded().Write(&b)
		conn.Write(b.Bytes())
		return
	}

	request := proxyHTTP.NewProxyRquest(
		req,
		s.config.ProxyID,
		s.pattern,
		conn.LocalAddr().(*net.TCPAddr),
		conn.RemoteAddr().(*net.TCPAddr),
	)
	res, err := s.client.Do(request.IntoForwarded())
	if err != nil {
		var proxyErr *errors.ProxyError
		if urlErr := err.(*url.Error); urlErr.Timeout() {
			proxyErr = errors.RequestTimeout()
		} else {
			proxyErr = errors.BadGateway()
		}

		res := proxyHTTP.ErrorToResponse(nil, proxyErr)
		b := bytes.Buffer{}
		res.IntoForwarded().Write(&b)
		conn.Write(b.Bytes())
		return
	}
	defer res.Body.Close()

	response := proxyHTTP.NewProxyResponse(res)
	b := bytes.Buffer{}
	response.IntoForwarded().Write(&b)
	conn.Write(b.Bytes())
}

func (s *Server) Run() {
	log.Printf("%s => Listening for requests", s.config.LogName)
	go func() {
	Loop:
		for {
			// TODO: Manejo de las maximas conexiones
			// if len(s.connections) == cap(s.connections) {
			// 	log.Println(
			// 		fmt.Sprintf(
			// 			"%s => Reached max connections: %d",
			// 			s.config.LogName, s.config.MaxConnections,
			// 		),
			// 	)
			// }

			conn, err := s.listener.AcceptTCP()
			if err != nil {
				break Loop
			}
			s.connections <- 1
			log.Printf(
				"%s => Accept new connection [%s]",
				s.config.LogName, conn.RemoteAddr().String(),
			)
			go s.forward(conn)
		}
	}()

	for {
		switch action := <-s.subscriber.Receiver; action {
		case notification.Report:
			s.subscriber.Sender <- notification.ServerState{}
		case notification.Shutdown:
			s.shutdown()
			return
		}
	}
}

func (s *Server) shutdown() {
	s.listener.Close()
	log.Printf("%s => Listening is closed", s.config.LogName)
	log.Printf(
		"%s => %d connections waiting to be closed",
		s.config.LogName, len(s.connections),
	)
	for {
		select {
		case action := <-s.subscriber.Receiver:
			if action == notification.Report {
				s.subscriber.Sender <- notification.ServerState{}
			}
		default:
			if len(s.connections) == 0 {
				log.Printf(
					"%s => All client connections have been closed ",
					s.config.LogName,
				)
				return
			}
		}
	}
}

func NewServer(config *config.Server, subscriber *notification.Subscriber) (*Server, error) {
	addr, err := net.ResolveTCPAddr("tcp", config.Listen)
	if err != nil {
		return nil, err
	}

	pattern, err := net.ResolveTCPAddr("tcp", config.Pattern)
	if err != nil {
		return nil, err
	}

	listener, err := net.ListenTCP(addr.Network(), addr)
	if err != nil {
		return nil, err
	}

	transport := http.Transport{
		Dial: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).Dial,
		MaxIdleConns:        config.MaxConnections,
		TLSHandshakeTimeout: 10 * time.Second,
	}
	client := http.Client{
		Transport: &transport,
		Timeout:   config.TimeoutPerRequest * time.Second,
	}

	connections := make(chan uint8, config.MaxConnections)
	return &Server{
		config:      config,
		subscriber:  subscriber,
		client:      &client,
		listener:    listener,
		pattern:     pattern,
		connections: connections,
	}, nil
}
