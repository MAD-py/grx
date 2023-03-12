package http

import (
	"bytes"
	"io"
	"net/http"
	"strings"

	"github.com/MAD-py/grx/pkg"
	"github.com/MAD-py/grx/pkg/errors"
)

type ProxyResponse struct {
	// Server response or proxy response in case of server or proxy error.
	response *http.Response
}

func (r *ProxyResponse) IntoForwarded() *http.Response {
	r.response.Header.Set("Server", pkg.Version())
	return r.response
}

func (r *ProxyResponse) CloseBody() {
	r.response.Body.Close()
}

func NewProxyResponse(res *http.Response) *ProxyResponse {
	return &ProxyResponse{
		response: res,
	}
}

func NewFileProxyResponse(req *http.Request, file []byte) *ProxyResponse {
	proto := req.Proto
	protoMajor := req.ProtoMajor
	protoMinor := req.ProtoMinor

	return &ProxyResponse{
		response: &http.Response{
			Status:     http.StatusText(http.StatusOK),
			StatusCode: http.StatusOK,

			Proto:      proto,
			ProtoMajor: protoMajor,
			ProtoMinor: protoMinor,

			Header: http.Header{},

			Body:          io.NopCloser(bytes.NewReader(file)),
			ContentLength: int64(len(file)),
		},
	}
}

// ErrorToResponse transforms an internal error into a processable http response,
// this function requires the original request from the client since it provides
// all the information of the protocol being used in the communication.
func ErrorToResponse(req *http.Request, err *errors.ProxyError) *ProxyResponse {
	proto := "HTTP/1.1"
	protoMajor := 1
	protoMinor := 1

	if req != nil {
		proto = req.Proto
		protoMajor = req.ProtoMajor
		protoMinor = req.ProtoMinor
	}

	return &ProxyResponse{
		response: &http.Response{
			Status:     http.StatusText(err.StatusCode()),
			StatusCode: err.StatusCode(),

			Proto:      proto,
			ProtoMajor: protoMajor,
			ProtoMinor: protoMinor,

			Header: http.Header{},

			Body:          io.NopCloser(strings.NewReader(err.Error())),
			ContentLength: int64(len(err.Error())),
		},
	}
}
