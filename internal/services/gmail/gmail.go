package gmail

import (
	"context"
	"time"

	"github.com/link00000000/gwsn/internal/services"
)

type GoogleApplicationConfig []byte

type Account struct {
	Username string
}

type gmailService struct {
	pollingInterval time.Duration
	appConfig       GoogleApplicationConfig
	accounts        []Account
}

var _ services.GmailService = (*gmailService)(nil)

func NewService(accounts []Account, appConfig GoogleApplicationConfig, pollingInterval time.Duration) *gmailService {
	return &gmailService{
		accounts:        accounts,
		appConfig:       appConfig,
		pollingInterval: pollingInterval,
	}
}

// implements [services.GmailService]
func (*gmailService) Setup() error {
	return nil
}

// implements [services.GmailService]
func (*gmailService) Run(ctx context.Context) error {
	return nil
}

// implements [services.GmailService]
func (*gmailService) Shutdown() error {
	return nil
}
