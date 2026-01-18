package googlecalendar

import (
	"context"

	"github.com/link00000000/gwsn/internal/services"
)

type ServiceConfig struct{}

type service struct {
	cfg ServiceConfig
}

var _ services.GoogleCalendarService = (*service)(nil)

func NewService(cfg ServiceConfig) *service {
	return &service{cfg: cfg}
}

// implements [services.GoogleCalendarService]
func (*service) Run(ctx context.Context) error {
	return nil
}
