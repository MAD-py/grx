package http

import (
	"fmt"
	"net"
	"net/http"
)

type ProxyRquest struct {
	// Original client request.
	request *http.Request

	// Proxy id used for the "by" field in the "Forwarded" header,
	// if the proxy does not have this id, the proxy id will be used.
	proxyID string

	// Address of the proxy
	proxyAddr *net.TCPAddr

	// Address of the client.
	clientAddr *net.TCPAddr

	// Address of the service processing this request
	forwardingAddr *net.TCPAddr
}

func (r *ProxyRquest) IntoForwarded() *http.Request {
	req := r.request.Clone(r.request.Context())
	req.URL.Host = r.forwardingAddr.String()
	req.URL.Scheme = "http"
	req.RequestURI = ""

	by := r.proxyAddr.String()
	if r.proxyID != "" {
		by = r.proxyID
	}

	forwarded := fmt.Sprintf(
		"for=%s;by=%s;host=%s",
		r.clientAddr.String(), by, r.request.Host,
	)
	if v := req.Header.Get("Forwarded"); v != "" {
		forwarded = fmt.Sprintf("%s, %s", v, forwarded)
	}
	req.Header.Set("Forwarded", forwarded)
	return req
}

func NewProxyRquest(
	req *http.Request,
	proxyID string,
	forwardingAddr, proxyAddr, clientAddr *net.TCPAddr,
) *ProxyRquest {
	return &ProxyRquest{
		request:        req,
		proxyID:        proxyID,
		proxyAddr:      proxyAddr,
		clientAddr:     clientAddr,
		forwardingAddr: forwardingAddr,
	}
}