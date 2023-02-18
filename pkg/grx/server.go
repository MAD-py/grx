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
)

const maxRequestSize = 32 << 20 // INFO: 32 MB

type server struct {
	// Configuration for the server.
	config *config.Server

	// Current server status.
	status serverStatus

	// HTTP client in charge of processing incoming requests .
	client *http.Client

	// TCP listener to accept incoming connections.
	listener *net.TCPListener

	// Forwarding TCP address
	pattern *net.TCPAddr

	// Connections are limited and this channel is used as a Semaphore
	// to prevent overloading.
	connections chan struct{}
}

func (s *server) shutdown() {
	if s.status == shuttingDown || s.status == offline {
		return
	}

	s.listener.Close()
	s.status = shuttingDown
	log.Printf("%s => Listening is closed", s.config.Name)
	log.Printf(
		"%s => %d connections waiting to be closed",
		s.config.Name, len(s.connections),
	)

	for {
		if len(s.connections) == 0 {
			log.Printf(
				"%s => All client connections have been closed ",
				s.config.Name,
			)
			s.status = offline
			return
		}
	}
}

func (s *server) forward(conn *net.TCPConn) {
	defer func() {
		conn.Close()
		log.Printf(
			"%s => Close connection [%s]",
			s.config.Name, conn.RemoteAddr().String(),
		)
		<-s.connections
	}()

	b := bytes.Buffer{}
	req, err := http.ReadRequest(bufio.NewReaderSize(conn, maxRequestSize))
	if err != nil {
		res := proxyHTTP.ErrorToResponse(nil, errors.BadRequest())
		res.IntoForwarded().Write(&b)
		conn.Write(b.Bytes())
		return
	}

	request := proxyHTTP.NewProxyRquest(
		req,
		s.config.ID,
		s.pattern,
		conn.LocalAddr().(*net.TCPAddr),
		conn.RemoteAddr().(*net.TCPAddr),
	)
	res, err := s.client.Do(request.IntoForwarded(true))
	if err != nil {
		var proxyErr *errors.ProxyError
		if urlErr := err.(*url.Error); urlErr.Timeout() {
			proxyErr = errors.RequestTimeout()
		} else {
			proxyErr = errors.BadGateway()
		}

		res := proxyHTTP.ErrorToResponse(req, proxyErr)
		res.IntoForwarded().Write(&b)
		conn.Write(b.Bytes())
		return
	}

	defer res.Body.Close()
	response := proxyHTTP.NewProxyResponse(res)
	response.IntoForwarded().Write(&b)
	conn.Write(b.Bytes())
}

func (s *server) run() {
	log.Printf("%s => Listening for requests", s.config.Name)
	s.status = online
Loop:
	for {
		conn, err := s.listener.AcceptTCP()
		if err != nil {
			break Loop
		}
		s.connections <- struct{}{}
		log.Printf(
			"%s => Accept new connection [%s]",
			s.config.Name, conn.RemoteAddr().String(),
		)
		go s.forward(conn)
	}
}

func newServer(config *config.Server) (*server, error) {
	addr, err := net.ResolveTCPAddr("tcp", config.ListenAddr)
	if err != nil {
		return nil, err
	}

	pattern, err := net.ResolveTCPAddr("tcp", config.PatternAddr)
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
	client := &http.Client{
		Transport: &transport,
		Timeout:   config.TimeoutPerRequest * time.Second,
	}

	connections := make(chan struct{}, config.MaxConnections)
	return &server{
		config:      config,
		status:      offline,
		client:      client,
		listener:    listener,
		pattern:     pattern,
		connections: connections,
	}, nil
}
