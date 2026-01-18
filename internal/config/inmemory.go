package config

import (
	"slices"
	"time"
)

type InMemoryConfigProvider struct {
	cfg *InMemoryConfig
}

var _ ConfigProvider = (*InMemoryConfigProvider)(nil)

type GmailAccountInMemoryConfig struct {
	Username *string
}

type GmailInMemoryConfig struct {
	Accounts        *[]GmailAccountInMemoryConfig
	PollingInterval *time.Duration
}

type InMemoryConfig struct {
	Gmail *GmailInMemoryConfig
}

func NewInMemoryConfigProvider(cfg *InMemoryConfig) *InMemoryConfigProvider {
	return &InMemoryConfigProvider{
		cfg: cfg,
	}
}

func (p *InMemoryConfigProvider) Apply(cfg *Config) error {
	if p.cfg.Gmail != nil {
		if p.cfg.Gmail.Accounts != nil {
			for _, acc := range *p.cfg.Gmail.Accounts {
				var targetAccount *GmailAccountConfig

				idx := slices.IndexFunc(cfg.Gmail.Accounts, func(a GmailAccountConfig) bool { return a.Username == *acc.Username })
				if idx == -1 {
					cfg.Gmail.Accounts = append(cfg.Gmail.Accounts, GmailAccountConfig{})
					targetAccount = &cfg.Gmail.Accounts[len(cfg.Gmail.Accounts)-1]
				} else {
					targetAccount = &cfg.Gmail.Accounts[idx]
				}

				applyProp(&targetAccount.Username, acc.Username)
			}
		}

		applyProp(&cfg.Gmail.PollingInterval, p.cfg.Gmail.PollingInterval)
	}

	return nil
}
