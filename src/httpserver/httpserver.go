package gwnhttp

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"
)

type Server struct {
	httpServer *http.Server
}

func NewGWNHttpServer() *Server {
	s := &Server{
		httpServer: &http.Server{
			Addr: ":8080",
		},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handle_root)
	mux.HandleFunc("/settings", s.handle_settings)

	s.httpServer.Handler = mux
	return s
}

func (s *Server) ListenAndServe(ctx context.Context) {
	log.Printf("GWN HTTP Server ListenAndServe started with Addr: %v\n", s.httpServer.Addr)

	go func() {
		s.httpServer.ListenAndServe()
	}()

	<-ctx.Done()

	log.Printf("Shutting down GWN HTTP Server")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	s.httpServer.Shutdown(shutdownCtx)
	cancel()
}

func (s *Server) handle_root(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "This this the root")
}

func (s *Server) handle_settings(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "This this the settings")
}
