package server

import (
	"fmt"
	"net/http"

	"github.com/user/patchwork/config"
)

// Server holds the HTTP server and its configuration.
type Server struct {
	cfg    *config.Config
	mux    *http.ServeMux
	server *http.Server
}

// New creates a new Server from the provided config, registering all routes.
func New(cfg *config.Config) *Server {
	mux := http.NewServeMux()

	for _, route := range cfg.Routes {
		handler := NewRouteHandler(route)
		mux.Handle(route.Path, handler)
	}

	handler := Chain(
		mux,
		LoggingMiddleware,
		CORSMiddleware,
	)

	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: handler,
	}

	return &Server{
		cfg:    cfg,
		mux:    mux,
		server: httpServer,
	}
}

// Start begins listening and serving HTTP requests.
func (s *Server) Start() error {
	fmt.Printf("patchwork listening on :%d\n", s.cfg.Port)
	return s.server.ListenAndServe()
}

// Handler returns the underlying http.Handler (useful for testing).
func (s *Server) Handler() http.Handler {
	return s.server.Handler
}
