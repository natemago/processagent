Process Agent
=============

Make everything a micro-service.

Process agent is a tool that can wrap an existing CLI app and expose it as micro-service.

It runs as an HTTP, WebSocket or AMQP endpoint, listens for requests on a specific port
and passes down the requests on the standard input of the wrapped process. Whatever it
receives from the process, returns back either as JSON (by default) or raw.

An example - run `date` as a microservice over HTTP:

```!bash
$ processagent -c "date" &
$ curl http://localhost:8080
{"id":"nqzKiMUepQMa","port":"http","payload":"Wed 10 Jul 2019 01:22:41 AM CEST\n","timestamp":1562714561667}
```

# Install

## Install from source

You can use Go to install it from source:

```bash
go get github.com/natemago/processagent
go install github.com/natemago/processagent
```

# Examples

## Simple service that reads from stdin and writes to stdout

A simple service that reads JSON from stdin, does dome processing and writes to
stdout:

File `service.go`:
```go
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

// Message holds the message data received as JSON.
type Message struct {
	Name string `json:"name"`
}

func main() {
	// Read all of the STDIN input.
	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		panic(err)
	}
	// Then unmarshal the JSON into our structure.
	message := &Message{}
	if err = json.Unmarshal(data, message); err != nil {
		panic(err)
	}

	// Print out the result on stdout.
	fmt.Printf("Hello %s! This is service.", message.Name)
}
```

Compile the file:
```bash
go build service.go
```

Then run it and call the service:

```bash
processagent -c "service" &
curl "http://localhost:8080" -d '{"name": "John Doe"}'
```

You will get the following result:
```
{"id":"5kHUO65OhkFH","port":"http","payload":"Hello John Doe! This is service.","timestamp":1563147018465}
```

## Run the processagent on different port

To run the processagent on different port, pass the `-p` parameter:

```bash
processagent -c "service" -p 8090
```

Then call it on that port:
```bash
curl "http://localhost:8090" -d '{"name": "John Doe"}'
```

and you should get the same result as above.


# What it is

Processagent is a simple tool designed to do a simple task of wrapping an existing
process and exposing it as a service on HTTP (and in future WebService and AMQP).
The communication between processagent and the wrapped process goes via the 
process `STDIN`, `STDOUT` and `STDERR`.

This design choice allows for maximum versatility of the wrapped processes, as it
is the simplest and most portable way of communication on all platforms.
The process itself can be written in any technology of your choosing, as long  as 
it can read from `STDIN` and write to `STDOUT` and `STDERR`. This is the case
with most of the current programming languages. On the other hand, the most popular
operating systems (GNU/Linux, Windows, Mac OSX and other *NIX systems) allow a 
process to read from stdin and write to stdout/stderr. Of course, some limitations do 
exist, usually on the size of the buffer for these streams, and are different on
different platforms. These must be taken into account.

There are many scenarios where a tool like this might be useful:

* You already have command-line tools that you use and want to expose as services.
These might be automation scripts in python, data analysis scripts in R or similar.
If your scripts already do what they need to do, and do it well, you don't want to
add complexity and possibly introduce bugs to them, as re-writing them as web services
might do so. Processagent will wrap the processes and won't add any complexity to your
code.
* You need to write services in many different technologies and want to abstract
the communication. All you need is to write the services to read from the process
stdin and write to stdout, without having to take care of the HTTP handling.
* You're desinging a large-scale system with different moving parts and processagent
is just a tool helping you to glue your modules as HTTP (or AMQP) services. This
is the original intention behind processagent.
* You're designing a FaaS distributed system. Processagent would handle the incoming
requests (over HTTP, Websocket, AMQP) and would pass the to the wrapped functions.
In this case, the wrapped processes can in fact be the functions that you want to
call.
* In data-center where you want to run data-intensive processing, and you want
to call the programs executing the algorithms as standard microservices.
* In a microservice architecture, you can have the absolute abstraction over
the technologies in which the microservices are implemented. Each microservice
can be implemented in any technology, without the added complexity of writing
the requests handling mechanism - just read from stdin, write to stdout.


# What it is not

* It is not a microservice framework. This is just a wrapper around a process 
that exposes that process onver HTTP, WebSocket or AMQP.
* It is not a REST framework. You can't define specifir routes or react to
different HTTP verbs - all requests get passed down to the wrapped process.
* It is not an AMQP or WebSocket client library.
* Does not support wrapping multiple commands. The command is given as cli argument
and cannot be chnaged on runtime. All arguments to the command are fixed and cannot
be chnaged also. This is due to security concerns as you dont want to give an
attacker option to run arbitrary processes.
* Does not support HTTP path matching (calling on /foo and /bar and /, do exactly
the same thing).

# Known issues and limitations

* Does not switch the user, nor is there a configuration for it. The processes are
run with the same user as processagent. For security reasons it is not reccomended
you run processagent with root or admin privileges, but with limited users only.
Best practice would be to run processagent with the same privileges that you would
run the undrlying process with.
* Does not support WebSockets (although this is on roadmap).
* Does not support AMQP (also on roadmap).

# How to contribute

First of all, thank you for taking interest in contributing on this project.

There are many ways you can contribute to this project. If you notice an issue or a behavior
that does not seems right or out-right crashes of the program, please submit an issue.
If you notice an error in the documentation, you can also submit an issue.
The tracker supports labels, so it would be nice if you can label the issues - bug,
enhancement etc.

If you already have a fix for some issue (bug fix, enhancement, documentation fix,
spelling error fix), you are more than welcomed to submit a pull request (PR).

To submit a PR, first clone this repository, do your fix in a branch on your
repository, and then submit a PR to merge your branch into the master branch of
this repository. Of course, make sure you have the latest changes from this
repository pulled into your branch, to avoid merge conflicts.

If you contribute code, it would be best to write unit tests for it - if it is a
bug fix, write a unit test showing that the bug have been fixed; if you write
an enhancement or new functionality, please cover that code with as much unit
test coverage as possible.

In the PR description, write down what that PR adds to the source code - does it
fix a bug, does it add some functionality and if so, what does it add. If you're
fixing an issue, it would be nice to reference it, so it may be closed once
the PR is accepted and merged.

PRs with malicious intent would not be accepted and will be closed.

