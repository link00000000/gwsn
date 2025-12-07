package systray

import (
	"context"
	"log"
	"sync/atomic"

	"github.com/getlantern/systray"
	"github.com/link00000000/google-workspace-notify/internal/systray/assets"
	"golang.org/x/sync/errgroup"
)

var running atomic.Bool

func init() {
	running.Store(false)
}

type Systray struct {
	g      *errgroup.Group
	ctx    context.Context
	cancel context.CancelFunc

	cExitReq chan struct{}
}

func NewSystray() *Systray {
	ctx, cancel := context.WithCancel(context.Background())
	g, ctx := errgroup.WithContext(ctx)

	return &Systray{
		g:        g,
		ctx:      ctx,
		cancel:   cancel,
		cExitReq: make(chan struct{}),
	}
}

func (s *Systray) Start() error {
	if !running.CompareAndSwap(false, true) {
		panic("attempted to start an instance of Systray while another one is already active")
	}

	s.g.Go(func() error {
		systray.Run(func() {
			systray.SetIcon(assets.TrayIcon)
			systray.SetTitle("Google Workspace Notify")

			mSettings := systray.AddMenuItem("Settings", "")
			s.g.Go(func() error { return s.runSystrayClickHandlerSettings(mSettings) })

			systray.AddSeparator()

			mExit := systray.AddMenuItem("Exit", "")
			s.g.Go(func() error { return s.runSystrayClickHandlerExit(mExit) })
		}, nil)

		return nil
	})

	return nil
}

func (s *Systray) Stop() error {
	if !running.CompareAndSwap(true, false) {
		panic("attempted to stop an instance of Systray but one is not active")
	}

	s.cancel()
	systray.Quit()
	if err := s.g.Wait(); err != nil {
		return err
	}

	return nil
}

func (s *Systray) ExitReq() <-chan struct{} {
	return s.cExitReq
}

func (s *Systray) runSystrayClickHandlerSettings(m *systray.MenuItem) error {
	for {
		select {
		case <-m.ClickedCh:
			log.Println("settings systray menu item clicked")
		case <-s.ctx.Done():
			return nil
		}
	}
}

func (s *Systray) runSystrayClickHandlerExit(m *systray.MenuItem) error {
	for {
		select {
		case <-m.ClickedCh:
			log.Println("exit systray menu item clicked")
			s.cExitReq <- struct{}{}
		case <-s.ctx.Done():
			return nil
		}
	}
}
