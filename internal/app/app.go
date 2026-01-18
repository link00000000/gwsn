package app

import (
	"context"
	"log/slog"

	"github.com/link00000000/gwsn/internal/services"
	"golang.org/x/sync/errgroup"
)

type shutdownRequest struct {
	RequestedByUser bool
	Reason          string
}

type ServiceContainer struct {
	gmail          services.GmailService
	googleCalendar services.GoogleCalendarService
	http           services.HttpService
	notification   services.NotificationService
	systemTray     services.SystemTrayService
}

func (svcs *ServiceContainer) runServices(ctx context.Context) error {
	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error { return svcs.gmail.Run(ctx) })
	g.Go(func() error { return svcs.googleCalendar.Run(ctx) })
	g.Go(func() error { return svcs.http.Run(ctx) })
	g.Go(func() error { return svcs.notification.Run(ctx) })
	g.Go(func() error { return svcs.systemTray.Run(ctx) })

	return g.Wait()
}

type Application struct {
	svcs             ServiceContainer
	logger           *slog.Logger
	shutdownRequests chan *shutdownRequest
}

var instance *Application = &Application{
	svcs:             ServiceContainer{},
	logger:           slog.Default(),
	shutdownRequests: make(chan *shutdownRequest),
}

func RegisterGmailService(svc services.GmailService) {
	instance.svcs.gmail = svc
}

func GmailService() services.GmailService {
	return instance.svcs.gmail
}

func RegisterGoogleCalendarService(svc services.GoogleCalendarService) {
	instance.svcs.googleCalendar = svc
}

func GoogleCalendarService() services.GoogleCalendarService {
	return instance.svcs.googleCalendar
}

func RegisterHttpService(svc services.HttpService) {
	instance.svcs.http = svc
}

func HttpService() services.HttpService {
	return instance.svcs.http
}

func RegisterNotificationService(svc services.NotificationService) {
	instance.svcs.notification = svc
}

func NotificationService() services.NotificationService {
	return instance.svcs.notification
}

func RegisterSystemTrayService(svc services.SystemTrayService) {
	instance.svcs.systemTray = svc
}

func SystemTrayService() services.SystemTrayService {
	return instance.svcs.systemTray
}

func ConfigureLogger(logger *slog.Logger) {
	instance.logger = logger
}

func Logger() *slog.Logger {
	return instance.logger
}

func Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error { return instance.svcs.runServices(ctx) })
	g.Go(func() error {
		select {
		case req := <-instance.shutdownRequests:
			Logger().Debug("received shutdown request", "requestedByUser", req.RequestedByUser, "reason", req.Reason)
			cancel()

		case <-ctx.Done():
		}

		return nil
	})

	return g.Wait()
}

func RequestShutdown(requestedByUser bool, reason string) {
	instance.shutdownRequests <- &shutdownRequest{requestedByUser, reason}
}
