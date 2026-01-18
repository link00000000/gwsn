package services

import (
	"context"
)

type Service interface {
	Setup() error // TODO: remove
	Run(ctx context.Context) error
	Shutdown() error // TODO: remove
}

type GmailService interface {
	Service
}

type GoogleCalendarService interface {
	Service
}

type HttpService interface {
	Service
}

type NotificationService interface {
	Service

	Notify(title, body string)
	NotifyWithIcon(title, body string, icon []byte)
}

type SystemTrayService interface {
	Service
}
