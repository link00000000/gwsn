package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/link00000000/gwsn/internal/app"
)

type GmailAccountConfig struct {
	Username string
}

type GmailConfig struct {
	Accounts        []GmailAccountConfig
	PollingInterval time.Duration
}

type Config struct {
	Gmail GmailConfig
}

type ConfigProvider interface {
	// Applies the provider options on top of the passed in cfg, overwriting.
	// If an error ocurrs, the provider's options are not applied and and error is returned.
	Apply(cfg *Config) error
}

// Builds a config using the list of providers. Providers are executed in order,
// each one overwriting the results of the previous.
func Build(providers ...ConfigProvider) (*Config, error) {
	cfg := &Config{
		Gmail: GmailConfig{
			Accounts: make([]GmailAccountConfig, 0),
		},
	}

	for _, p := range providers {
		if err := p.Apply(cfg); err != nil {
			app.Logger().Info("config skipped due to previous error")
		}

	}

	app.Logger().Debug("Finished building config", "config", cfg)

	return cfg, nil
}

type filePathType string

const (
	filePathType_UserConfig filePathType = "UserConfig"
	filePathType_Cwd        filePathType = "CurrentWorkingDirectory"
	filePathType_Literal    filePathType = "Literal"
)

type filePath struct {
	_type filePathType
	name  string
}

func LiteralFilePath(name string) *filePath {
	return &filePath{_type: filePathType_Literal, name: name}
}

func CwdRelFilePath(name string) *filePath {
	return &filePath{_type: filePathType_Cwd, name: name}
}

func UserConfigRelFilePath(name string) *filePath {
	return &filePath{_type: filePathType_UserConfig, name: name}
}

func (f *filePath) Resolve() (string, error) {
	switch f._type {
	case filePathType_UserConfig:
		dir, err := os.UserConfigDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(dir, "gwsn", f.name), nil
	case filePathType_Cwd:
		return filepath.Abs(f.name)
	case filePathType_Literal:
		return f.name, nil
	default:
		panic(fmt.Sprintf("unexpected config.filePathType: %#v", f._type))
	}
}

func applyProp[T any](target *T, source *T) bool {
	if source != nil {
		*target = *source
		return true
	}

	return false
}
