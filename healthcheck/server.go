package healthcheck

import (
	"context"
	"github.com/ONSdigital/go-ns/log"
	"github.com/ONSdigital/go-ns/server"
	"github.com/gorilla/mux"
	"time"
)

// Server provides a HTTP server specifically for health checks.
// Use this if you only require a HTTP server for health checks.
type Server struct {
	httpServer *server.Server
	ticker     *Ticker
}

func NewServer(bindAddr string, duration time.Duration, errorChannel chan error, clients ...Client) *Server {

	router := mux.NewRouter()
	router.Path("/healthcheck").HandlerFunc(Do)

	ticker := NewTicker(duration, clients...)

	httpServer := server.New(bindAddr, router)

	// Disable auto handling of os signals by the HTTP server. This is handled
	// in the service so we can gracefully shutdown resources other than just
	// the HTTP server.
	httpServer.HandleOSSignals = false

	go func() {
		log.Debug("starting http server", log.Data{"bind_addr": bindAddr})
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
