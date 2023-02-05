package errors

import "net/http"

type ProxyError struct {
	text       string
	statusCode int
}

func (e ProxyError) Error() string {
	return e.text
}

func (e ProxyError) StatusCode() int {
	return e.statusCode
}

// ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓ //
// ┃               Client error              ┃ //
// ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛ //

func BadRequest() *ProxyError {
	return &ProxyError{
		text:       "HTTP 400 BAD REQUEST",
		statusCode: http.StatusBadRequest,
	}
}

func NotFound() *ProxyError {
	return &ProxyError{
		text:       "HTTP 404 NOT FOUND",
		statusCode: http.StatusNotFound,
	}
}

func RequestTimeout() *ProxyError {
	return &ProxyError{
		text:       "HTTP 408 REQUEST TIMEOUT",
		statusCode: http.StatusRequestTimeout,
	}
}

// ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓ //
// ┃               Server error              ┃ //
// ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛ //

func BadGateway() *ProxyError {
	return &ProxyError{
		text:       "HTTP 502 BAD GATEWAY",
		statusCode: http.StatusBadGateway,
	}
}
