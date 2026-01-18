package config

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"slices"
	"time"

	"github.com/link00000000/gwsn/internal/app"
)

type JSONDuration time.Duration

var _ json.Unmarshaler = (*JSONDuration)(nil)

func (d *JSONDuration) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	dur, err := time.ParseDuration(s)
	if err != nil {
		return err
	}

	*d = JSONDuration(dur)
	return nil
}

type gmailAccountJsonConfig struct {
	Username *string `json:"username"`
}

type gmailJsonConfig struct {
	Accounts        *[]gmailAccountJsonConfig `json:"accounts"`
	PollingInterval *JSONDuration             `json:"pollingIntervalSeconds"`
}

type jsonConfig struct {
	Gmail *gmailJsonConfig `json:"gmail"`
}

type JsonConfigProvider struct {
	path *filePath
}

var _ ConfigProvider = (*JsonConfigProvider)(nil)

func NewJsonFileConfigProvider(path *filePath) *JsonConfigProvider {
	return &JsonConfigProvider{
		path: path,
	}
}

func (p *JsonConfigProvider) Apply(cfg *Config) error {
	name, err := p.path.Resolve()
	if err != nil {
		app.Logger().Error("failed to resolve config file path", "error", err, "config_name", p.path.name, "config_type", p.path._type)
		return err
	}

	b, err := os.ReadFile(name)
	if errors.Is(err, fs.ErrNotExist) {
		app.Logger().Warn("config file does not exist, skipping", "config_name", p.path.name, "config_type", p.path._type, "resolved_config_name", name)
		return nil
	}

	if err != nil {
		app.Logger().Error("failed to read JSON config file", "error", err, "config_name", p.path.name, "config_type", p.path._type, "resolved_config_name", name)
		return err
	}

	jsonCfg := jsonConfig{}
	if err := json.Unmarshal(b, &jsonCfg); err != nil {
		app.Logger().Error("failed to parse JSON config file", "error", err, "config_name", p.path.name, "config_type", p.path._type, "resolved_config_name", name)
		return err
	}

	if jsonCfg.Gmail != nil {
		if jsonCfg.Gmail.Accounts != nil {
			for _, acc := range *jsonCfg.Gmail.Accounts {
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

		applyProp(&cfg.Gmail.PollingInterval, (*time.Duration)(jsonCfg.Gmail.PollingInterval))
	}

	return nil
}
