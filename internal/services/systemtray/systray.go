package systemtray

import (
	"context"

	"github.com/getlantern/systray"
	"github.com/link00000000/gwsn/internal/app"
	"github.com/link00000000/gwsn/internal/services"
)

type ServiceConfig struct {
	Title    string
	TrayIcon []byte
}

type service struct {
	cfg ServiceConfig
}

var _ services.SystemTrayService = (*service)(nil)

func NewSystraySystemTrayService(cfg ServiceConfig) *service {
	return &service{cfg: cfg}
}

// implements [services.SystemTrayService]
func (svc *service) Run(ctx context.Context) error {
	systray.Run(func() {
		systray.SetIcon(svc.cfg.TrayIcon)
		systray.SetTitle(svc.cfg.Title)

		systray.AddSeparator()
		exitEntry := systray.AddMenuItem("Exit", "")

		for {
			select {
			case <-exitEntry.ClickedCh:
				app.RequestShutdown(true, "system tray exit menu entry clicked")

			case <-ctx.Done():
				systray.Quit()
				return
			}
		}
	}, nil)

	return nil
}
