package main

import (
	"context"
	_ "embed"
	"log/slog"
	"os"
	"time"

	"github.com/link00000000/gwsn/internal/app"
	"github.com/link00000000/gwsn/internal/config"
	"github.com/link00000000/gwsn/internal/services/gmail"
	"github.com/link00000000/gwsn/internal/services/googlecalendar"
	"github.com/link00000000/gwsn/internal/services/http"
	"github.com/link00000000/gwsn/internal/services/notification"
	"github.com/link00000000/gwsn/internal/services/systemtray"
	"github.com/link00000000/gwsn/internal/services/systemtray/assets"
)

const (
	AppName = "Google Workspace Notify"
)

var (
	DefaultGmailPollingInterval = time.Minute * 5

	DefaultConfig = config.InMemoryConfig{
		Gmail: &config.GmailInMemoryConfig{
			PollingInterval: &DefaultGmailPollingInterval,
		},
	}

	// go:embed credentials.json
	appConfig gmail.GoogleApplicationConfig
)

func main() {
	app.ConfigureLogger(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})))

	cfg, err := config.Build(
		config.NewInMemoryConfigProvider(&DefaultConfig),
		config.NewJsonFileConfigProvider(config.UserConfigRelFilePath("config.json")),
		config.NewJsonFileConfigProvider(config.CwdRelFilePath("config.json")),
	)

	if err != nil {
		app.Logger().Error("failed to build config", "error", err)
		os.Exit(1)
	}

	// Gmail service
	gmailConfig := gmail.ServiceConfig{}
	gmailConfig.AppConfig = appConfig
	gmailConfig.PollingInterval = cfg.Gmail.PollingInterval
	for _, a := range cfg.Gmail.Accounts {
		gmailConfig.Accounts = append(gmailConfig.Accounts, gmail.ServiceConfigAccount{Username: a.Username})
	}
	app.RegisterGmailService(gmail.NewService(gmailConfig))

	// Google calendar service
	googleCalendarConfig := googlecalendar.ServiceConfig{}
	app.RegisterGoogleCalendarService(googlecalendar.NewService(googleCalendarConfig))

	// Http service
	httpConfig := http.ServiceConfig{}
	httpConfig.Addr = "127.0.0.1:8080"
	httpConfig.Routes = []http.ServiceConfigRoute{}
	app.RegisterHttpService(http.NewService(httpConfig))

	// Notification service
	notificationConfig := notification.ServiceConfig{}
	notificationConfig.AppName = AppName
	app.RegisterNotificationService(notification.NewBeeepNotificationService(notificationConfig))

	// System tray service
	systemTrayConfig := systemtray.ServiceConfig{}
	systemTrayConfig.Title = AppName
	systemTrayConfig.TrayIcon = assets.TrayIcon
	app.RegisterSystemTrayService(systemtray.NewSystraySystemTrayService(systemTrayConfig))

	if err := app.Run(context.Background()); err != nil {
		panic(err)
	}
}
