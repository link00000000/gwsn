package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/link00000000/google-workspace-notify/internal/gworkspace"
	"github.com/link00000000/google-workspace-notify/internal/sysnotif"
	"github.com/link00000000/google-workspace-notify/internal/systray"
	"github.com/link00000000/google-workspace-notify/internal/ui"
	"golang.org/x/sync/errgroup"
)

func RunSystray(ctx context.Context, cancel context.CancelFunc) error {
	s := systray.NewSystray()
	s.Start()

loop:
	for {
		select {
		case <-s.ExitReq():
			cancel()
		case <-ctx.Done():
			break loop
		}
	}

	if err := s.Stop(); err != nil {
		return err
	}

	return nil
}

func RunMonitor(ctx context.Context) error {
	m, err := gworkspace.NewMonitor()
	if err != nil {
		return fmt.Errorf("failed to create monitor: %v", err)
	}

	go m.Run() // TODO: Handle error and early terminate
	defer m.Stop()

	for {
		select {
		case <-m.CalendarReminder():
			slog.Info("recieved calendar reminder") // TODO: add attrs
			sysnotif.ShowNotification("Upcoming calendar event", "There is an upcoming calendar event")
			// TODO: notify new calendar reminder
		case <-m.Email():
			slog.Info("recieved email") // TODO: add attrs
			sysnotif.ShowNotification("New email", "There is a new email")
			// TODO: notify new email
		case <-ctx.Done():
			return nil
		}
	}
}

func RunHttpServer(ctx context.Context) error {
	s := &http.Server{Addr: ":8080", Handler: ui.NewHandler()}

	go s.ListenAndServe()
	<-ctx.Done()

	return s.Shutdown(context.TODO())
}

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})))

	ctx, cancel := context.WithCancel(context.Background())
	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		slog.Info("starting systray")

		err := RunSystray(ctx, cancel)
		if err != nil {
			panic(fmt.Errorf("RunSystray completed with unhandled error: %v", err))
		}

		slog.Info("systray completed without error")

		return nil
	})

	g.Go(func() error {
		slog.Info("starting RunHttpServer")

		err := RunHttpServer(ctx)
		if err != nil {
			panic(fmt.Errorf("RunHttpServer completed with unhandled error: %v", err))
		}

		slog.Info("RunHttpServer completed without error")

		return nil
	})

	g.Go(func() error {
		slog.Info("starting RunMonitor")

		err := RunMonitor(ctx)
		if err != nil {
			panic(fmt.Errorf("RunMonitor completed with unhandled error: %v", err))
		}

		slog.Info("RunMonitor completed without error")

		return nil
	})

	if err := g.Wait(); err != nil {
		panic(err)
	}
}
