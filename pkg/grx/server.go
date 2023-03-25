package grx

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/MAD-py/grx/pkg/config"
	"github.com/MAD-py/grx/pkg/errors"

	proxyHTTP "github.com/MAD-py/grx/pkg/http"
)

type server interface {
	run()
	shutdown()

	forward(*net.TCPConn)
	getStatus() serverStatus
}

const maxRequestSize = 32 << 20 // INFO: 32 MB

type baseServer struct {
	// Name of the server that will be visible in the logs
	name string

	// Current server status.
	status serverStatus

	// TCP listener to accept incoming connections.
	listener *net.TCPListener

	// Connections are limited and this channel is used as a Semaphore
	// to prevent overloading.
	connections chan struct{}
}

func (s *baseServer) getStatus() serverStatus { return s.status }

func (s *baseServer) shutdown() {
	if s.status == shuttingDown || s.status == offline {
		return
	}

	s.listener.Close()
	s.status = shuttingDown
	log.Printf("%s => Listening is closed", s.name)
	log.Printf(
		"%s => %d connections waiting to be closed",
		s.name, len(s.connections),
	)

	for {
		if len(s.connections) == 0 {
			log.Printf(
				"%s => All client connections have been closed",
				s.name,
			)
			s.status = offline
			return
		}
	}
}

type forwardServer struct {
	baseServer

	// Proxy id used for the "by" field in the "Forwarded" header.
	id string

	// Value to select between the two types of available headers
	useForwarded bool

	// HTTP client in charge of processing incoming requests.
	client *http.Client

	// Forwarding TCP address.
	pattern string
}

func (s *forwardServer) forward(conn *net.TCPConn) {
	defer func() {
		conn.Close()
		log.Printf(
			"%s => Close connection [%s]",
			s.name, conn.RemoteAddr().String(),
		)
		<-s.connections
	}()

	b := bytes.Buffer{}
	req, err := http.ReadRequest(bufio.NewReaderSize(conn, maxRequestSize))
	if err != nil {
		res := proxyHTTP.ErrorToResponse(nil, errors.BadRequest())
		res.IntoForwarded().Write(&b)
		conn.Write(b.Bytes())
		res.CloseBody()
		return
	}

	request := proxyHTTP.NewProxyRquest(
		req,
		s.id,
		s.pattern,
		conn.LocalAddr().String(),
		conn.RemoteAddr().String(),
	)
	res, err := s.client.Do(request.IntoForwarded(s.useForwarded))
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
		res.CloseBody()
		return
	}

	response := proxyHTTP.NewProxyResponse(res)
	response.IntoForwarded().Write(&b)
	conn.Write(b.Bytes())
	response.CloseBody()
}

func (s *forwardServer) run() {
	log.Printf("Starting the forward server %s", s.name)
	log.Printf("%s => Listening for requests", s.name)
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
			s.name, conn.RemoteAddr().String(),
		)
		go s.forward(conn)
	}
}

type staticServer struct {
	baseServer

	// Prefix to complete static routes
	pathPrefix string
}

func (s *staticServer) forward(conn *net.TCPConn) {
	defer func() {
		conn.Close()
		log.Printf(
			"%s => Close connection [%s]",
			s.name, conn.RemoteAddr().String(),
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

	path := filepath.Join(s.pathPrefix, req.URL.Path)
	file, err := os.ReadFile(path)
	if err != nil {
		res := proxyHTTP.ErrorToResponse(nil, errors.NotFound())
		res.IntoForwarded().Write(&b)
		conn.Write(b.Bytes())
		res.CloseBody()
		return
	}

	res := proxyHTTP.NewFileProxyResponse(req, file)
	res.IntoForwarded().Write(&b)
	conn.Write(b.Bytes())
	res.CloseBody()
}

func (s *staticServer) run() {
	log.Printf("Starting the static server %s", s.name)
	log.Printf("%s => Listening for requests", s.name)
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
			s.name, conn.RemoteAddr().String(),
		)
		go s.forward(conn)
	}
}

func newForwardServer(config *config.ForwardServer) (*forwardServer, error) {
	addr, err := net.ResolveTCPAddr("tcp", config.ListenAddr)
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

	return &forwardServer{
		baseServer: baseServer{
			name:        config.Name,
			status:      offline,
			listener:    listener,
			connections: make(chan struct{}, config.MaxConnections),
		},
		id:           config.ID,
		client:       client,
		pattern:      config.PatternAddr,
		useForwarded: config.UseForwarded,
	}, nil
}

func newStaticServer(config *config.StaticServer) (*staticServer, error) {
	if _, err := os.Stat(config.PathPrefix); os.IsNotExist(err) {
		return nil, fmt.Errorf("the %s folder does not exist", config.PathPrefix)
	}

	addr, err := net.ResolveTCPAddr("tcp", config.ListenAddr)
	if err != nil {
		return nil, err
	}

	listener, err := net.ListenTCP(addr.Network(), addr)
	if err != nil {
		return nil, err
	}

	return &staticServer{
		baseServer: baseServer{
			name:        config.Name,
			status:      offline,
			listener:    listener,
			connections: make(chan struct{}, config.MaxConnections),
		},
		pathPrefix: config.PathPrefix,
	}, nil
}
