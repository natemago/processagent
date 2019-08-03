package processagent

import (
	"context"
)

// InputPort represents a point of entry of the incoming requests to be processed.
// An input port may be for example an HTTP listener, WebSocket server or AMQP
// topic or queue.
// An incoming message (HTTP request, message over WebSocker or consumed from
// AMQP queue) is consumed from the appropriate port, then parsed. A Response
// object is generated and the incoming Request together with the generated
// Response and Context are processed by the middlewares registered in this
// input port.
type InputPort interface {

	// AddMiddleware adds a middleware to the list of middlewares for this port.
	// Every request (message) is processed by all middlewares registered in the
	// chain for this port.
	AddMiddleware(middleware Middleware)

	// Close shuts down this port.
	// This operation is synchronous - waits for process of port shutdown to
	// complete. If this fails, it returns an error.
	Close() error
}

// MiddlewareInputPort holds a list of middlewares comprising the chain for this
// port.
// It implements the basic middleware chain management and is intended to be
// embedded in a specific implementations of the InputPort interface.
// It offers a special method for triggering the execution of all middlewares in
// the chain, called ExecuteMiddlewares.
type MiddlewareInputPort struct {
	middlewares []Middleware
}

// AddMiddleware adds a Middleware to this endpoint.
func (m *MiddlewareInputPort) AddMiddleware(middleware Middleware) {
	m.middlewares = append(m.middlewares, middleware)
}

// Close shuts down the input port. This implementation does nothing.
func (m *MiddlewareInputPort) Close() error {
	return nil
}

// ExecuteMiddlewares executes the middleware chain with the given context, Request and Response.
// If any of the middlewares in the chain produces an error, the chain is broken and the error is
// returned.
func (m *MiddlewareInputPort) ExecuteMiddlewares(ctx context.Context, req *Request, resp *Response) error {
	for _, middleware := range m.middlewares {
		if err := middleware(ctx, req, resp); err != nil {
			return err
		}
	}
	return nil
}

// NewMiddlewarePort creates new MiddlewareInputPort and initializes the middleware chain.
func NewMiddlewarePort() *MiddlewareInputPort {
	return &MiddlewareInputPort{
		middlewares: []Middleware{},
	}
}
