package server

import (
	"fmt"
	"log"
	"net/http"

	"github.com/user/patchwork/config"
)

// Server holds the HTTP server and its configuration.
type Server struct {
	cfg    *config.Config
	mux    *http.ServeMux
	httpSrv *http.Server
}

// New creates a new Server from the given config.
func New(cfg *config.Config) *Server {
	mux := http.NewServeMux()
	s := &Server{
		cfg: cfg,
		mux: mux,
		httpSrv: &http.Server{
			Addr:    fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
			Handler: mux,
		},
	}
	s.registerRoutes()
	return s
}

// registerRoutes registers all routes defined in the config.
func (s *Server) registerRoutes() {
	for _, route := range s.cfg.Routes {
		r := route // capture loop variable
		pattern := fmt.Sprintf("%s %s", r.Method, r.Path)
		s.mux.HandleFunc(pattern, NewRouteHandler(r))
		log.Printf("registered route: %s %s -> %d", r.Method, r.Path, r.Response.Status)
	}
}

// Start begins listening for HTTP requests.
func (s *Server) Start() error {
	log.Printf("patchwork listening on %s", s.httpSrv.Addr)
	return s.httpSrv.ListenAndServe()
}
