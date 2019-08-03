package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	pa "github.com/natemago/processagent"
)

type configuredPorts []pa.InputPort

func (p *configuredPorts) AddPort(port pa.InputPort) {
	*p = append(*p, port)
}

func (p *configuredPorts) AddMiddleware(middleware pa.Middleware) {
	for _, port := range *p {
		port.AddMiddleware(middleware)
	}
}

func (p *configuredPorts) Close() {
	for _, port := range *p {
		if err := port.Close(); err != nil {
			log.Println("Failed to close port: ", err.Error())
		}
	}
}

func main() {
	if err := pa.RunCLI(func(cfg *pa.Config) error {
		ports := &configuredPorts{}

		// configure ports
		ports.AddPort(pa.NewHTTPEndpoint("", *cfg.Port, "/"))

		// run process agent
		processAgent := pa.NewProcessAgent(*cfg.Command, *cfg.MaxWorkers)

		// configure middlewares
		worker := processAgent.GetMiddleware()

		for _, handler := range []pa.Handler{pa.ResponseTimestamp, pa.JSONResponse, pa.RequestID(9), pa.RequestTimestamp} {
			worker = handler(worker)
		}

		ports.AddMiddleware(worker)

		done := make(chan bool)
		c := make(chan os.Signal)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-c
			processAgent.Stop()
			ports.Close()
			done <- true
		}()
		<-done
		return nil
	}); err != nil {
		log.Fatal(err)
	}
}
