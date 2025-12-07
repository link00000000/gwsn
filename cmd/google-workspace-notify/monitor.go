package main

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/link00000000/google-workspace-notify/src/monitor"
)

func RunMonitor(ctx context.Context) error {
	m, err := monitor.NewMonitor()
	if err != nil {
		return fmt.Errorf("failed to create monitor: %v", err)
	}

	go m.Run() // TODO: Handle error and early terminate
	defer m.Stop()

	for {
		select {
		case <-m.CalendarReminder():
			slog.Info("recieved calendar reminder") // TODO: add attrs
			ShowNotification("Upcoming calendar event", "There is an upcoming calendar event")
			// TODO: notify new calendar reminder
		case <-m.Email():
			slog.Info("recieved email") // TODO: add attrs
			ShowNotification("New email", "There is a new email")
			// TODO: notify new email
		case <-ctx.Done():
			return nil
		}
	}
}
