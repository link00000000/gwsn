package services

import (
	"context"
)

type Service interface {
	Run(ctx context.Context) error
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
