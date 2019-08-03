package processagent

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"testing"
)

func TestGenerateRandomString(t *testing.T) {
	str := GenerateRandomString(10)
	if str == "" {
		log.Fatal("Expected random string to be generated.")
	}
}

func TestCurrentTimeMillis(t *testing.T) {
	now := CurrentTimeMillis()
	if now <= 0 {
		t.Fatal("Current time cannot be 0 or less")
	}
}

func TestRequestID(t *testing.T) {
	reqID := RequestID(12)
	middleware := func(ctx context.Context, req *Request, resp *Response) error {
		if req.ID == "" {
			return fmt.Errorf("request id not set")
		}
		return nil
	}
	middleware = reqID(middleware)

	if err := middleware(context.Background(), &Request{}, &Response{}); err != nil {
		t.Fatal(err)
	}
}

func TestRequestTimestamp(t *testing.T) {
	middleware := func(ctx context.Context, req *Request, resp *Response) error {
		if req.Timestamp <= 0 {
			return fmt.Errorf("request timestamp invalid")
		}
		return nil
	}

	middleware = RequestTimestamp(middleware)
	if err := middleware(context.Background(), &Request{}, &Response{}); err != nil {
		t.Fatal(err)
	}
}

func TestResponseTimestamop(t *testing.T) {
	middleware := func(ctx context.Context, req *Request, resp *Response) error {
		return nil
	}
	middleware = ResponseTimestamp(middleware)
	resp := &Response{}
	middleware(context.Background(), &Request{}, resp)
	if resp.Timestamp <= 0 {
		t.Fatal("Response timestamp not set properly")
	}
}

func TestJSONResponse(t *testing.T) {
	middleware := func(ctx context.Context, req *Request, resp *Response) error {
		return nil
	}
	middleware = JSONResponse(middleware)
	resp := &Response{
		ID:        "test-id",
		Timestamp: 1000,
	}
	middleware(context.Background(), &Request{}, resp)
	if resp.Payload == "" {
		t.Fatal("Expected the Payload of the response to be set.")
	}

	result := &Response{}
	if err := json.Unmarshal([]byte(resp.Payload), result); err != nil {
		t.Fatal(err)
	}
	if result.ID != resp.ID {
		t.Fatal("Response was not serialized properly. ID mismatch.")
	}
	if result.Timestamp != resp.Timestamp {
		t.Fatal("Response was not serialized properly. Timestamp mismatch.")
	}
}
