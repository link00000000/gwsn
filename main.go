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
	gmailAccounts := make([]gmail.Account, len(cfg.Gmail.Accounts))
	for i, a := range cfg.Gmail.Accounts {
		gmailAccounts[i].Name = a.Username
	}

	app.RegisterGmailService(gmail.NewService(cfg.Gmail.PollingInterval, gmailAccounts))

	// Google calendar service
	app.RegisterGoogleCalendarService(googlecalendar.NewService())

	// Notification service
	app.RegisterNotificationService(notification.NewBeeepNotificationService(AppName))

	// System tray service
	app.RegisterSystemTrayService(systemtray.NewSystraySystemTrayService(AppName, assets.TrayIcon))

	if err := app.Run(context.Background()); err != nil {
		panic(err)
	}
}
