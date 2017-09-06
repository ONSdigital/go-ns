package server

import (
	"net/http"
	"os"
	"os/signal"
	"time"

	"context"
	"github.com/ONSdigital/go-ns/handlers/requestID"
	"github.com/ONSdigital/go-ns/handlers/timeout"
	"github.com/ONSdigital/go-ns/log"
	"github.com/justinas/alice"
)

// Server is a http.Server with sensible defaults, which supports
// configurable middleware and timeouts, and shuts down cleanly
// on SIGINT/SIGTERM
type Server struct {
	http.Server
	Middleware             map[string]alice.Constructor
	MiddlewareOrder        []string
	Alice                  *alice.Chain
	CertFile               string
	KeyFile                string
	DefaultShutdownTimeout time.Duration
	HandleOSSignals        bool
}

// New creates a new server
func New(bindAddr string, router http.Handler) *Server {
	middleware := map[string]alice.Constructor{
		"RequestID": requestID.Handler(16),
		"Log":       log.Handler,
		"Timeout":   timeout.Handler(10 * time.Second),
	}

	return &Server{
		Alice:           nil,
		Middleware:      middleware,
		MiddlewareOrder: []string{"RequestID", "Log", "Timeout"},
		Server: http.Server{
			Handler:           router,
			Addr:              bindAddr,
			ReadTimeout:       5 * time.Second,
			WriteTimeout:      10 * time.Second,
			ReadHeaderTimeout: 0,
			IdleTimeout:       0,
			MaxHeaderBytes:    0,
		},
		HandleOSSignals:        true,
		DefaultShutdownTimeout: 10 * time.Second,
	}
}

func (s *Server) prep() {
	var m []alice.Constructor
	for _, v := range s.MiddlewareOrder {
		if mw, ok := s.Middleware[v]; ok {
			m = append(m, mw)
			continue
		}
		panic("middleware not found: " + v)
	}

	s.Server.Handler = alice.New(m...).Then(s.Handler)
}

// ListenAndServe sets up SIGINT/SIGTERM signals, builds the middleware
// chain, and creates/starts a http.Server instance
//
// If CertFile/KeyFile are both set, the http.Server instance is started
// using ListenAndServeTLS. Otherwise ListenAndServe is used.
//
// Specifying one of CertFile/KeyFile without the other will panic.
func (s *Server) ListenAndServe() error {
	if s.HandleOSSignals {
		return s.listenAndServeHandleOSSignals()
	}

	return s.listenAndServe()
}

// ListenAndServeTLS sets KeyFile and CertFile, then calls ListenAndServe
func (s *Server) ListenAndServeTLS(certFile, keyFile string) error {
	if len(certFile) == 0 || len(keyFile) == 0 {
		panic("either CertFile/KeyFile must be blank, or both provided")
	}
	s.KeyFile = keyFile
	s.CertFile = certFile
	return s.ListenAndServe()
}

// Shutdown will gracefully shutdown the server, using a default shutdown
// timeout if a context is not provided.
func (s *Server) Shutdown(ctx context.Context) error {

	if ctx == nil {
		ctx, _ = context.WithTimeout(context.Background(), s.DefaultShutdownTimeout)
	}

	return s.Server.Shutdown(ctx)
}

func (s *Server) listenAndServe() error {

	s.prep()
	if len(s.CertFile) > 0 || len(s.KeyFile) > 0 {
		return s.Server.ListenAndServeTLS(s.CertFile, s.KeyFile)
	}

	return s.Server.ListenAndServe()
}

func (s *Server) listenAndServeHandleOSSignals() error {

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, os.Kill)

	s.listenAndServeAsync()

	<-stop
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	return s.Shutdown(ctx)
}

func (s *Server) listenAndServeAsync() {

	s.prep()
	if len(s.CertFile) > 0 || len(s.KeyFile) > 0 {
		go func() {
			if err := s.Server.ListenAndServeTLS(s.CertFile, s.KeyFile); err != nil {
				log.Error(err, nil)
				os.Exit(1)
			}
		}()
	} else {
		go func() {
			if err := s.Server.ListenAndServe(); err != nil {
				log.Error(err, nil)
				os.Exit(1)
			}
		}()
	}
}
