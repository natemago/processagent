package processagent

import (
	"context"
	"testing"
)

func TestInputPortAddMiddleware(t *testing.T) {
	endpoint := &MiddlewareInputPort{
		middlewares: []Middleware{},
	}

	endpoint.AddMiddleware(func(ctx context.Context, req *Request, resp *Response) error {
		return nil
	})

	if len(endpoint.middlewares) != 1 {
		t.Fatal("Expected to have 1 middleware but instead got ", len(endpoint.middlewares))
	}

	if err := endpoint.Close(); err != nil {
		t.Fatal("Expected the port to close without error. Error: ", err.Error())
	}
}

func TestExecuteMiddlewares(t *testing.T) {
	executed := &[]string{}
	middleware := func(name string) Middleware {
		return func(ctx context.Context, req *Request, r *Response) error {
			*executed = append(*executed, name)
			return nil
		}
	}

	port := &MiddlewareInputPort{
		middlewares: []Middleware{
			middleware("one"),
			middleware("two"),
			middleware("three"),
		},
	}

	if err := port.ExecuteMiddlewares(context.Background(), &Request{}, &Response{}); err != nil {
		t.Fatal("Expected not to get error while executing middlewares. Error: ", err.Error())
	}

	if len(*executed) != 3 {
		t.Fatal("Expected all three middlewares to have executed.")
	}

	if (*executed)[0] != "one" || (*executed)[1] != "two" || (*executed)[2] != "three" {
		t.Fatal("Middlewares did no execute in the expected order. Actual order is: ", *executed)
	}
}

func TestNewMiddlewarePort(t *testing.T) {
	port := NewMiddlewarePort()
	if port == nil {
		t.Fatal("Expected a pointer to MiddlewareInputPort.")
	}
}
