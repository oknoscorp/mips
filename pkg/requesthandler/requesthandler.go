package requesthandler

import (
	"net"
)

type RequestHandler struct {
	Requests  chan *Request
	Responses chan *Response
}

type Request struct {
	Conn  *net.Conn
	Input string
}

type Response struct {
	Conn    *net.Conn   `json:"-"`
	Code    string      `json:"code"`
	Content interface{} `json:"content"`
	Error   error       `json:"-"`
}

// New will initialize request handler and create
// buffered channels.
func New() *RequestHandler {
	return &RequestHandler{
		Requests:  make(chan *Request),
		Responses: make(chan *Response),
	}
}
