package notification

import (
	"context"

	"github.com/gen2brain/beeep"
	"github.com/link00000000/gwsn/internal/services"
)

type ServiceConfig struct {
	AppName string
}

type service struct {
	cfg ServiceConfig
}

var _ services.NotificationService = (*service)(nil)

func NewBeeepNotificationService(cfg ServiceConfig) *service {
	return &service{cfg: cfg}
}

// implements [services.NotificationService]
func (svc *service) Run(ctx context.Context) error {
	beeep.AppName = svc.cfg.AppName

	return nil
}

// implements [services.NotificationService]
func (*service) Notify(title, body string) {
	beeep.Notify(title, body, "")
}

// implements [services.NotificationService]
func (*service) NotifyWithIcon(title, body string, icon []byte) {
	beeep.Notify(title, body, icon)
}
