package processagent

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os/exec"
	"strings"
	"sync"
	"syscall"
)

// ProcessAgent defines an interface for the external processes execution and
// management.
// It executes a new process, then passes the given Request to the external
// process and populates the given Response.
type ProcessAgent interface {
	// Stop shuts down the process agent and terminates all running processes.
	// The call is synchronous, meaning it returns after all processes have
	// terminated successfully or with error.
	Stop() error

	// ProcessCommand runs a new process, passing the given Request to that
	// process. Once the process completes, the Response is populated with the
	// result of the process or, in case of an error, an error is returned.
	// The processing of the request is syncrhonous and the function runs until
	// the processing is complete.
	ProcessCommand(req *Request, resp *Response) (err error)
}

// processEvent defines a function that reacts to process events, such as
// process start and process end.
type processEvent func(pw *processWrapper)

// processWrapper holds the configuration for a single process run.
type processWrapper struct {
	cmd           *exec.Cmd
	stdin         io.Reader
	stdout        *bytes.Buffer
	stderr        *bytes.Buffer
	processStarts processEvent
	processEnds   processEvent
	running       bool
}

// runProcess runs a single process. The executable is specified by execStr and
// the Request is passed down to the external process on STDIN of the process.
// The execStr is tokenized into arguments, of which the first is the executable
// and the rest (if any) are passed as arguments to the process.
func (w *processWrapper) runProcess(req *Request, execStr string) (string, error) {
	execStr = strings.TrimSpace(execStr)
	if execStr == "" {
		return "", fmt.Errorf("no exec specified")
	}
	args, err := Tokenize(execStr)
	if err != nil {
		return "", err
	}
	executable := args[0]
	if len(args) > 1 {
		args = args[1:]
	} else {
		args = []string{}
	}

	outStr, errStr := w.exec(req.Payload, executable, args)
	if errStr != "" {
		return "", fmt.Errorf(errStr)
	}

	return outStr, nil
}

// callEnd is called when the external process terminates.
func (w *processWrapper) callEnd() {
	if w.running {
		w.running = false
		if w.processEnds != nil {
			w.processEnds(w)
		}
	}
}

// exec executes an external process. The process command is specified via
// the executable parameter and any arguments are passed via args.
// An additional input is passed down to the process via the STDIN on the
// external process.
// The function returns whatever the external process prints on the STDOUT and
// STDERR.
func (w *processWrapper) exec(input string, executable string, args []string) (outStr, errStr string) {
	if w.running {
		return "", "already running"
	}
	w.running = true
	w.cmd = exec.Command(executable, args...)
	w.stdin = strings.NewReader(input)
	w.cmd.Stdin = w.stdin
	w.cmd.Stdout = w.stdout
	w.cmd.Stderr = w.stderr

	defer func() {
		w.callEnd()
	}()

	if err := w.cmd.Start(); err != nil {
		return "", err.Error()
	}

	if w.processStarts != nil {
		go w.processStarts(w)
	}

	if err := w.cmd.Wait(); err != nil {
		return "", err.Error()
	}

	errStr = w.stderr.String()
	if errStr != "" {
		return "", errStr
	}

	outStr = w.stdout.String()

	return outStr, errStr
}

// stopProcess terminates the external process. The process is signaled with
// SIGTERM to terminate gracefully.
func (w *processWrapper) stopProcess() error {
	defer func() {
		w.callEnd()
	}()
	if !w.running {
		// don't try to stop the process
		return nil
	}
	if err := w.cmd.Process.Signal(syscall.SIGTERM); err != nil {
		return nil
	}

	_, err := w.cmd.Process.Wait()
	if err != nil {
		return err
	}
	return nil
}

// newProcessWrapper creates new process wrapper with the given callback handlers
// for process start and process end.
// onStart is called right after the process starts and everything is set up.
// onEnd is called when the process terminates, either successfully or with
// an error.
func newProcessWrapper(onStart, onEnd processEvent) *processWrapper {
	return &processWrapper{
		stderr:        &bytes.Buffer{},
		stdout:        &bytes.Buffer{},
		processEnds:   onEnd,
		processStarts: onStart,
	}
}

// LocalProcessAgent holds the configuration for running local processes.
// It always runs the same executable (the same command) configurable via
// execCommand field.
// If maxParallel is specified (not 0), then it limits the number of processes
// running at the same time to this number.
type LocalProcessAgent struct {
	execCommand string
	maxParallel int
	running     map[int]*processWrapper
	lock        sync.Mutex
}

// GetMiddleware returns a middleware that can be attached to a given InputPort
// to handle Request by running a local process with this process agent.
func (p *LocalProcessAgent) GetMiddleware() Middleware {
	return func(ctx context.Context, req *Request, resp *Response) error {
		return p.ProcessCommand(req, resp)
	}
}

// Stop shuts down all currently running processes.
func (p *LocalProcessAgent) Stop() error {
	for pid, pw := range p.running {
		if err := pw.stopProcess(); err != nil {
			log.Printf("Process with pid %d failed to stop: %s\n", pid, err.Error())
		}
	}
	return nil
}

// ProcessCommand handles a Request by running a new process.
// If maxParallel is set, and the maximal number of currently running processes
// is reached, then the call would return an error immediately.
func (p *LocalProcessAgent) ProcessCommand(req *Request, resp *Response) error {
	if p.maxParallel != 0 && p.maxParallel <= len(p.running) {
		return fmt.Errorf("max number of workers reached")
	}

	pw := newProcessWrapper(func(pw *processWrapper) {
		p.lock.Lock()
		p.running[pw.cmd.Process.Pid] = pw
		p.lock.Unlock()

	}, func(pw *processWrapper) {
		if pw.cmd.Process != nil {
			p.lock.Lock()
			delete(p.running, pw.cmd.Process.Pid)
			p.lock.Unlock()
		}
	})

	output, err := pw.runProcess(req, p.execCommand)
	resp.Payload = output

	if err != nil {
		errv := true
		errCode := 500
		resp.Error = &errv
		resp.ErrorCode = &errCode
		resp.Payload = err.Error()
		log.Println("ProcessAgent: Failed to process command. Error:", err.Error())
	}
	return nil
}

// NewProcessAgent creates and configures new LocalProcessAgent with the given
// executable command.
// The max number of processes that can be run simultaneously is set by maxParallel.
// For unlimited number of simultaneous processes set this parameter to 0.
func NewProcessAgent(execCommand string, maxParallel int) *LocalProcessAgent {
	return &LocalProcessAgent{
		execCommand: execCommand,
		maxParallel: maxParallel,
		running:     map[int]*processWrapper{},
	}
}

// Tokenize parses an input command line string (like a bash/shell command) into
// an array of arguments similarly like bash does.
// For example: "ls -la my-dir" would yield ["ls", "-la", "my-dir"].
// Similarly: "echo \"test\"" would yield ["echo", "test"].
// String interpolation is not performed.
func Tokenize(str string) ([]string, error) {
	tokens := []string{}

	token := ""
	i := 0

	strGroup := false
	for {
		if i == len(str) {
			break
		}
		c := str[i]
		if c == '\\' {
			if i == len(str)-1 {
				break
			}
			next := str[i+1]
			if next == '\\' || next == '"' || next == '\'' {
				c = next
				i++
			}
			token += string(c)
			i++
		} else if c == '"' || c == '\'' {
			if strGroup {
				tokens = append(tokens, token)
				token = ""
				strGroup = false
				i++
				continue
			} else {
				strGroup = true
				i++
				continue
			}
		} else if c == ' ' || c == '\t' || c == '\n' || c == '\r' {
			if strGroup {
				token += string(c)
				i++
			} else {
				if token != "" {
					tokens = append(tokens, token)
					token = ""
				}
				i++
			}
		} else {
			token += string(c)
			i++
		}
	}
	if token != "" {
		if strGroup {
			return nil, fmt.Errorf("unclosed string group")
		}
		tokens = append(tokens, token)
	}

	return tokens, nil
}
