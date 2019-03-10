package healthcheck

import (
	"context"
	"time"

	"github.com/ONSdigital/go-ns/server"
	"github.com/ONSdigital/log.go/log"
	"github.com/gorilla/mux"
)

// Server provides a HTTP server specifically for health checks.
// Use this if you only require a HTTP server for health checks.
type Server struct {
	httpServer *server.Server
	ticker     *Ticker
}

func NewServer(bindAddr string, duration, recoveryDuration time.Duration, errorChannel chan error, clients ...Client) *Server {
	return NewServerWithAlerts(bindAddr, duration, recoveryDuration, errorChannel, nil, nil, clients...)
}

func NewServerWithAlerts(bindAddr string, duration, recoveryDuration time.Duration, errorChannel chan error, stateChangeChan, requestCheckChan chan bool, clients ...Client) *Server {
	router := mux.NewRouter()
	router.Path("/healthcheck").HandlerFunc(Do)

	ticker := NewTickerWithAlerts(duration, recoveryDuration, stateChangeChan, requestCheckChan, clients...)

	httpServer := server.New(bindAddr, router)

	// Disable auto handling of os signals by the HTTP server. This is handled
	// in the service so we can gracefully shutdown resources other than just
	// the HTTP server.
	httpServer.HandleOSSignals = false

	go func() {
		log.Event(nil, "starting http server", log.Data{"bind_addr": bindAddr})
		if err := httpServer.ListenAndServe(); err != nil {
			if errorChannel != nil {
				errorChannel <- err
			}
		}
	}()

	return &Server{
		httpServer: httpServer,
		ticker:     ticker,
	}
}

func (server *Server) Close(ctx context.Context) error {
	server.ticker.Close()
	return server.httpServer.Close(ctx)
}
