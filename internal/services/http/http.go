package http

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/link00000000/gwsn/internal/services"
)

type HttpServiceRoute struct {
	PathPrefix string
	Handler    http.Handler
}

type HttpServiceConfig struct {
	Addr   string
	Routes []HttpServiceRoute
}

type httpService struct {
	server *http.Server
}

var _ services.HttpService = (*httpService)(nil)

func NewService(cfg HttpServiceConfig) *httpService {
	mux := http.NewServeMux()
	for _, r := range cfg.Routes {
		mux.Handle(r.PathPrefix, http.StripPrefix(r.PathPrefix, r.Handler))
	}

	server := &http.Server{
		Addr:    cfg.Addr,
		Handler: mux,
	}

	return &httpService{server: server}
}

func (*httpService) Setup() error {
	return nil
}

func (svc *httpService) Run(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		err := svc.server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := svc.server.Shutdown(shutdownCtx)
	err = errors.Join(err, <-errCh)

	return err
}

func (*httpService) Shutdown() error {
	return nil
}
