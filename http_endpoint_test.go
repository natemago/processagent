package processagent

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewHTTPEndpoint(t *testing.T) {
	httpEndpoint := NewHTTPEndpoint("", 10113, "/")

	if httpEndpoint == nil {
		t.Fatal("Expected valid pointer to an HTTP Input Port")
	}

	if err := httpEndpoint.Close(); err != nil {
		t.Fatal("Failed to close the HTTP Port correctly.")
	}
}

func TestHttpEndpointMiddleware(t *testing.T) {
	middlewareCalled := false
	done := make(chan bool)
	httpEndpoint := NewHTTPEndpoint("", 10113, "/a")
	httpEndpoint.AddMiddleware(func(ctx context.Context, req *Request, resp *Response) error {
		middlewareCalled = true
		if req == nil {
			t.Fatal("Request must be present.")
		}
		if req.Payload != "TEST" {
			t.Fatal("Request payload is incorrect.")
		}

		if resp == nil {
			t.Fatal("Response is not provided.")
		}

		resp.Payload = "TEST-RESPONSE"

		go func() { done <- true }()
		return nil
	})

	req := httptest.NewRequest("POST", "/a", strings.NewReader("TEST"))
	resp := httptest.NewRecorder()

	http.DefaultServeMux.ServeHTTP(resp, req)
	go func() {
		time.Sleep(time.Duration(5) * time.Second)
		done <- true
	}()
	<-done
	if !middlewareCalled {
		t.Fatal("Expected the middleware to be called.")
	}

	if resp.Body.String() != "TEST-RESPONSE" {
		t.Fatal("Response payload is invalid. Expected \"TEST-RESPONSE\", but instead got: ", resp.Body.String())
	}

}
