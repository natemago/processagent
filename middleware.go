package processagent

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"log"
	"time"
)

// Request wraps the incoming request from an input port.
type Request struct {
	// ID is the unique ID of this request.
	ID string `json:"id"`
	// Port type of port on which this request was initially received.
	Port string `json:"port"`
	// Payload holds the original request payload.
	Payload string `json:"payload"`
	// Timestamp is the Unix timestamp (in milliseconds) when the request was
	// received.
	Timestamp int64 `json:"timestamp"`
}

// Response represents a response to a particular Request.
type Response struct {
	// ID is the response ID. It matches the Request ID.
	ID string `json:"id"`
	// Port is the port type on which the request was received and to which this
	// response is going to be send.
	Port string `json:"port"`
	// Payload the response payload as string.
	Payload string `json:"payload"`
	// Timestamp is the Unix timestamp (in milliseconds) when this response was
	// ready to be send back.
	Timestamp int64 `json:"timestamp"`
	// Error if not nil, then signals that an error occurred.
	Error *bool `json:"error,omitempty"`
	// ErrorCode is the code of the error. Used in hinting the actual error code
	// for the specific port. Present only if Error is set to true.
	ErrorCode *int `json:"errorCode,omitempty"`
}

// Middleware is a function called for every Request received on a particular
// endpoint.
// Each middleware executes in a given Context.
// A pointer to a pre-generated Response wrapper is also passed as a parameter
// to each middleware.
type Middleware func(context.Context, *Request, *Response) error

// Handler is a decorator type for Middleware.
// A Handler wraps the given Middleware and returns the middleware wrapper.
// This gives the possibility to easily extend and decorate an exiting middleware
// by chaining. For example, a handler might be defined like so:
//	func GenerateRequestID(middleware Middleware) Middleware {
//		return func(ctx context.Context, req *Request, resp *Response) error {
//			if req.ID == "" {
//				req.ID = generateRandomID()
//			}
//			return middleware(ctx, req, response)
//		}
//	}
type Handler func(Middleware) Middleware

// GenerateRandomString generates random string in base64 encoding by generating
// a random bytes of the given byteSize and then encoding the bytes.
func GenerateRandomString(byteSize int) string {
	buff := make([]byte, byteSize)
	if _, err := rand.Read(buff); err != nil {
		log.Fatal("Failed to generate random string: ", err.Error())
	}

	return base64.StdEncoding.EncodeToString(buff)
}

// CurrentTimeMillis returns a UNIX timestamp in milliseconds.
func CurrentTimeMillis() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

// RequestID is a Handler that generates a random ID for the Request.
func RequestID(size int) Handler {
	return func(middleware Middleware) Middleware {
		return func(ctx context.Context, req *Request, resp *Response) error {
			id := GenerateRandomString(size)
			req.ID = id
			resp.ID = id
			return middleware(ctx, req, resp)
		}
	}
}

// RequestTimestamp is a Handler that adds timestamp (in millisecods) to the
// middleware Request.
func RequestTimestamp(middleware Middleware) Middleware {
	return func(ctx context.Context, req *Request, resp *Response) error {
		now := CurrentTimeMillis()
		req.Timestamp = now
		return middleware(ctx, req, resp)
	}
}

// ResponseTimestamp is a Handler that adds timestamp (in millisecods) to the
// middleware Response.
func ResponseTimestamp(middleware Middleware) Middleware {
	return func(ctx context.Context, req *Request, resp *Response) error {
		err := middleware(ctx, req, resp)
		resp.Timestamp = CurrentTimeMillis()
		return err
	}
}

// JSONResponse is a Handler that serializes the whole Response as JSON and
// sets it as a Payload of the Response. Note that this overwrites the value
// of the Payload in the Response.
// The serialization executes after the original middleware has executed.
func JSONResponse(middleware Middleware) Middleware {
	return func(ctx context.Context, req *Request, resp *Response) error {
		err := middleware(ctx, req, resp)
		if err != nil {
			return err
		}
		data, err := json.Marshal(resp)
		if err != nil {
			return err
		}
		resp.Payload = string(data)
		return nil
	}
}
