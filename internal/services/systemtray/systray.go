package systemtray

import (
	"context"

	"github.com/getlantern/systray"
	"github.com/link00000000/gwsn/internal/app"
	"github.com/link00000000/gwsn/internal/services"
)

type systraySystemTrayService struct {
	title    string
	trayIcon []byte
}

var _ services.SystemTrayService = (*systraySystemTrayService)(nil)

func NewSystraySystemTrayService(title string, trayIcon []byte) *systraySystemTrayService {
	return &systraySystemTrayService{
		title:    title,
		trayIcon: trayIcon,
	}
}

// implements [services.SystemTrayService]
func (svc *systraySystemTrayService) Run(ctx context.Context) error {
	systray.Run(func() {
		systray.SetIcon(svc.trayIcon)
		systray.SetTitle(svc.title)

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
