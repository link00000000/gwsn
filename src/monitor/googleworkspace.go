package monitor

import (
	"log/slog"
	"time"
)

type Monitor struct {
	cCalendarReminder chan CalendarReminder
	cEmail            chan Email

	stop chan struct{}
}

type CalendarReminder struct {
}

type Email struct {
}

func NewMonitor() *Monitor {
	return &Monitor{
		cCalendarReminder: make(chan CalendarReminder, 32),
		cEmail:            make(chan Email, 32),
		stop:              make(chan struct{}),
	}
}

func (m *Monitor) CalendarReminder() <-chan CalendarReminder {
	return m.cCalendarReminder
}

func (m *Monitor) Email() <-chan Email {
	return m.cEmail
}

func (m *Monitor) Run() {
	t := time.NewTicker(time.Second * 5)
	for {
		select {
		case <-t.C:
			m.pollCalendarReminders()
			m.pollEmails()
		case <-m.stop:
			close(m.cCalendarReminder)
			close(m.cEmail)
			return
		}
	}
}

func (m *Monitor) Stop() {
	close(m.stop)
}

func (m *Monitor) pollCalendarReminders() {
	slog.Debug("polling calendar reminders")
	// TODO

	m.cCalendarReminder <- CalendarReminder{}
}

func (m *Monitor) pollEmails() {
	slog.Debug("polling emails")
	// TODO

	m.cEmail <- Email{}
}
