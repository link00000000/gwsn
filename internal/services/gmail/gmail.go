package gmail

import (
	"context"
	"time"

	"github.com/link00000000/gwsn/internal/services"
)

type GoogleApplicationConfig []byte

type ServiceConfigAccount struct {
	Username string
}

type ServiceConfig struct {
	Accounts        []ServiceConfigAccount
	AppConfig       GoogleApplicationConfig
	PollingInterval time.Duration
}

type service struct {
	cfg ServiceConfig
}

var _ services.GmailService = (*service)(nil)

func NewService(cfg ServiceConfig) *service {
	return &service{cfg: cfg}
}

// implements [services.GmailService]
func (*service) Run(ctx context.Context) error {
	return nil
}
