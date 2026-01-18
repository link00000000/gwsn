package http

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/link00000000/gwsn/internal/services"
)

type ServiceConfigRoute struct {
	PathPrefix string
	Handler    http.Handler
}

type ServiceConfig struct {
	Addr   string
	Routes []ServiceConfigRoute
}

type service struct {
	cfg ServiceConfig
}

var _ services.HttpService = (*service)(nil)

func NewService(cfg ServiceConfig) *service {
	return &service{cfg: cfg}
}

// implements [services.HttpService]
func (svc *service) Run(ctx context.Context) error {
	mux := http.NewServeMux()
	for _, r := range svc.cfg.Routes {
		mux.Handle(r.PathPrefix, http.StripPrefix(r.PathPrefix, r.Handler))
	}

	server := &http.Server{
		Addr:    svc.cfg.Addr,
		Handler: mux,
	}

	errCh := make(chan error, 1)
	go func() {
		err := server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := server.Shutdown(shutdownCtx)
	err = errors.Join(err, <-errCh)

	return err
}
