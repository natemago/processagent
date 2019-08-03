package processagent

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestTokenize(t *testing.T) {

	tokens, err := Tokenize("ls -la ./")
	if err != nil {
		t.Fatal("Failed to parse simple args", err)
	}
	if tokens == nil {
		t.Fatal("Expected tokens to be parsed.")
	}
	if len(tokens) != 3 {
		t.Fatal("expected exactly 3 tokens but got", len(tokens))
	}

	tokens, err = Tokenize("ls \"-la ./\"")
	if err != nil {
		t.Fatal("Failed to parse command line with string quotation", err)
	}
	if tokens == nil {
		t.Fatal("Expected tokens to be parsed.")
	}

	if len(tokens) != 2 {
		t.Fatal("expected exactly 2 tokens but got", len(tokens))
	}

	tokens, err = Tokenize("/bin/sh -c \"echo \\\"Hello there, General Kenobi\\\"\"")
	if err != nil {
		t.Fatal("Failed to parse command line with string quotation", err)
	}
	if tokens == nil {
		t.Fatal("Expected tokens to be parsed.")
	}
	if len(tokens) != 3 {
		t.Fatal("expected exactly 3 tokens but got", len(tokens))
	}
	fmt.Println(strings.Join(tokens, "|"))
}

func TestNewProcessWrapper(t *testing.T) {
	processStartExecuted := false
	processEndExecuted := false
	pw := newProcessWrapper(func(p *processWrapper) {
		if p == nil {
			t.Fatal("Expected pointer to processWrapper in process start")
		}
		processStartExecuted = true
	}, func(p *processWrapper) {
		if p == nil {
			t.Fatal("Expected pointer to processWrapper in process end")
		}
		processEndExecuted = true
	})

	out, err := pw.exec("", "/bin/sh", []string{"-c", "echo \"test\""})
	if err != "" {
		t.Fatal("Expected no error, but got:", err)
	}
	if out != "test\n" {
		t.Fatal("Expected to get \"test\" as an output, but got:", out)
	}

	if !processStartExecuted {
		t.Fatal("Process start callback not executed")
	}
	if !processEndExecuted {
		t.Fatal("Process end callback not executed")
	}
}

func TestProcessWrapperRunProcess(t *testing.T) {
	pw := newProcessWrapper(nil, nil)

	out, err := pw.runProcess(&Request{
		Payload: "test",
	}, "/bin/sh -c \"cat\"")

	if err != nil {
		t.Fatal(err)
	}
	if out != "test" {
		t.Fatal("Expected to get \"test\" as output, but instead got:", out)
	}
}

func TestProcessWrapperStopProcess(t *testing.T) {
	done := false
	pw := newProcessWrapper(nil, func(p *processWrapper) {
		done = true
	})

	go func() {
		time.Sleep(time.Duration(5) * time.Second)
		if !done {
			t.Fatal("Should have been terminated, but still running.")
		}
	}()
	go func() {
		time.Sleep(time.Duration(2) * time.Second)
		err := pw.stopProcess()
		if err != nil {
			t.Fatal("Failed to stop process. Error:", err.Error())
		}
	}()
	pw.runProcess(&Request{
		Payload: "",
	}, "/bin/sh -c \"sleep 30\"")
}

func TestProcessAgentStartThenStop(t *testing.T) {
	pa := NewProcessAgent("/bin/sh -c \"echo 'test'\"", 0)
	if err := pa.Stop(); err != nil {
		t.Fatal(err)
	}
}

func TestProcessAgentRunMiddleware(t *testing.T) {
	pa := NewProcessAgent("/bin/sh -c \"cat\"", 0)
	paMiddleware := pa.GetMiddleware()

	resp := &Response{}
	if err := paMiddleware(context.Background(), &Request{
		Payload: "test",
	}, resp); err != nil {
		t.Fatal(err)
	}

	if resp.Payload != "test" {
		t.Fatal("Expected to get 'test' as response payload, but instead got:", resp.Payload)
	}
}

func TestProcessingAgentRunMultipleRequests(t *testing.T) {
	pa := NewProcessAgent("/bin/sh -c \"sleep 2\"", 3)
	paMiddleware := pa.GetMiddleware()

	allDone := false
	done := make(chan bool)

	go func() {
		time.Sleep(time.Duration(4) * time.Second)
		if !allDone {
			t.Fatal("Expected for all middlewares to be done.")
		}
	}()

	for i := 0; i < 3; i++ {
		go func() {
			if err := paMiddleware(context.Background(), &Request{}, &Response{}); err != nil {
				t.Fatal("Got an error, but expected to run normally. Error:", err.Error())
			}
			done <- true
		}()
	}

	for i := 0; i < 3; i++ {
		<-done
	}
	allDone = true

}

func TestProcessAgentRejectRequest(t *testing.T) {
	pa := NewProcessAgent("/bin/sh -c \"sleep 2\"", 3)
	paMiddleware := pa.GetMiddleware()

	allDone := false
	done := make(chan bool)

	go func() {
		time.Sleep(time.Duration(4) * time.Second)
		if !allDone {
			t.Fatal("Expected for all middlewares to be done.")
		}
	}()

	for i := 0; i < 3; i++ {
		go func() {
			if err := paMiddleware(context.Background(), &Request{}, &Response{}); err != nil {
				t.Fatal("Got an error, but expected to run normally. Error:", err.Error())
			}
			done <- true
		}()
	}

	// pause a little to give time for the middlewares to run
	time.Sleep(time.Duration(500) * time.Millisecond)

	err := paMiddleware(context.Background(), &Request{}, &Response{})
	if err == nil {
		defer t.Fatal("Expected the request to be denied as all worker slots are occupied.")
	}

	if err != nil && err.Error() != "max number of workers reached" {
		defer t.Fatal("Expected the request to be reject but instead got error: ", err.Error())
	}

	for i := 0; i < 3; i++ {
		<-done
	}
	allDone = true

	// if all good, now we can schedule request as all middlewares completed
	if err = paMiddleware(context.Background(), &Request{}, &Response{}); err != nil {
		t.Fatal("Expected middleware to be able to run, but got an error:", err.Error())
	}
}
