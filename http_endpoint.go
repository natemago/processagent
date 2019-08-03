package processagent

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

// HTTPEndpoint represents an InputPort that handles HTTP requests.
// Wraps an HTTP server (see http.Server) that handles the HTTP requests.
type HTTPEndpoint struct {
	InputPort *MiddlewareInputPort
	Server    http.Server
}

// AddMiddleware adds a Middleware to the http input port.
func (h *HTTPEndpoint) AddMiddleware(middleware Middleware) {
	h.InputPort.AddMiddleware(middleware)
}

// Close shuts down the underlying HTTP server and closes this input port.
func (h *HTTPEndpoint) Close() error {
	return h.Server.Shutdown(context.Background())
}

// handleHTTPRequest is an http.Handler and handles a single HTTP request.
// This function maps the incoming HTTP requests, creates the Request and Response
// structures for the middleware chain, then executes the registered middlewares.
func (h *HTTPEndpoint) handleHTTPRequest(rw http.ResponseWriter, req *http.Request) {
	payloadData, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Println("HTTP Port: Failed to read request body: ", err.Error())
		return
	}

	requestWrapper := &Request{
		Port:    "http",
		Payload: string(payloadData),
	}

	ctx := context.Background()

	resp := &Response{
		Port: "http",
	}

	if err = h.InputPort.ExecuteMiddlewares(ctx, requestWrapper, resp); err != nil {
		log.Println("HTTP Port: Failed to process request: ", err.Error())
		return
	}

	statusCode := 200
	if resp.Error != nil && *resp.Error {
		statusCode = 500
		if resp.ErrorCode != nil {
			statusCode = *resp.ErrorCode
		}
	}

	rw.WriteHeader(statusCode)
	rw.Write([]byte(resp.Payload))
}

// NewHTTPEndpoint creates new HTTP InputPort starting an HTTP Server that
// listens on the given host and port. The port only handles requests comming on
// the given path pattern. To handle all requests provide "/" as a pattern.
func NewHTTPEndpoint(host string, port int, pattern string) *HTTPEndpoint {
	endpoint := &HTTPEndpoint{
		Server: http.Server{
			Addr: fmt.Sprintf("%s:%d", host, port),
		},
		InputPort: NewMiddlewarePort(),
	}

	http.HandleFunc(pattern, endpoint.handleHTTPRequest)

	go func() {
		if err := endpoint.Server.ListenAndServe(); err != nil {
			log.Println("Http Server: ", err.Error())
		}
	}()

	return endpoint
}
