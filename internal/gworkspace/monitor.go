package gworkspace

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"golang.org/x/sync/errgroup"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

const (
	credentialsFilePath = "credentials.json"
	tokenFilePath       = "token.json"
)

type Monitor struct {
	cCalendarReminder chan *CalendarReminder
	cEmail            chan *cachedEmail

	stop chan struct{}

	ctx    context.Context
	cancel context.CancelFunc

	httpClient *http.Client
	calSvc     *calendar.Service
	gmailSvc   *gmail.Service
}

type CalendarReminder struct {
	eventName string
	eventDesc string
	time      time.Time
}

func NewMonitor() (*Monitor, error) {
	ctx, cancel := context.WithCancel(context.Background())

	m := &Monitor{
		cCalendarReminder: make(chan *CalendarReminder, 32),
		cEmail:            make(chan *cachedEmail, 32),
		stop:              make(chan struct{}),
		ctx:               ctx,
		cancel:            cancel,
	}

	b, err := os.ReadFile(credentialsFilePath)
	if err != nil {
		// TODO: Handle better
		return nil, err
	}

	cfg, err := google.ConfigFromJSON(b, gmail.GmailReadonlyScope, calendar.CalendarReadonlyScope)
	if err != nil {
		return nil, fmt.Errorf("failed to get oauth config: %v", err)
	}

	tok, err := getToken(m.ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to get oauth token: %v", err)
	}

	m.httpClient = cfg.Client(m.ctx, tok)

	m.calSvc, err = calendar.NewService(m.ctx, option.WithHTTPClient(m.httpClient))
	if err != nil {
		return nil, fmt.Errorf("failed to create calendar service", "error", err)
	}

	m.gmailSvc, err = gmail.NewService(m.ctx, option.WithHTTPClient(m.httpClient))
	if err != nil {
		return nil, fmt.Errorf("failed to create gmail service", "error", err)
	}

	return m, nil
}

func (m *Monitor) CalendarReminder() <-chan *CalendarReminder {
	return m.cCalendarReminder
}

func (m *Monitor) Email() <-chan *cachedEmail {
	return m.cEmail
}

func (m *Monitor) Run() error {
	m.poll(m.ctx)

	t := time.NewTicker(time.Minute * 5)
	for {
		select {
		case <-t.C:
			m.poll(m.ctx)
		case <-m.stop:
			close(m.cCalendarReminder)
			close(m.cEmail)
			return nil
		}
	}
}

func getCachedToken() (*oauth2.Token, error) {
	slog.Debug("getting cached token from file", "file", tokenFilePath)

	f, err := os.Open(tokenFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open token file (%s): %v", tokenFilePath, err)
	}
	defer f.Close()

	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)

	if err != nil {
		return nil, fmt.Errorf("failed to parse token file (%s): %v", tokenFilePath, err)
	}

	return tok, nil
}

func setCachedToken(token *oauth2.Token) error {
	slog.Debug("caching token in file", "file", tokenFilePath)

	f, err := os.OpenFile(tokenFilePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to open token file (%s): %v", tokenFilePath, err)
	}
	defer f.Close()

	err = json.NewEncoder(f).Encode(token)
	if err != nil {
		return fmt.Errorf("failed to write token file (%s): %v", tokenFilePath, err)
	}

	return nil
}

func getToken(ctx context.Context, cfg *oauth2.Config) (*oauth2.Token, error) {
	tok, err := getCachedToken()
	if err != nil {
		slog.Warn("failed to get cached token", "file", tokenFilePath, "error", err)

		tok, err = getTokenFromWeb(ctx, cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to get token new token: %v", err)
		}

		err = setCachedToken(tok)
		if err != nil {
			// This error is okay because we will just get a new token next time
			slog.Error("failed to set cached token", "error", err)
		}
	}
	return tok, nil
}

func getTokenFromWeb(ctx context.Context, cfg *oauth2.Config) (*oauth2.Token, error) {
	slog.Info("getting new token from the web")

	stateToken := "state-token" // TODO: Generate proper state token
	authUrl := cfg.AuthCodeURL(stateToken, oauth2.AccessTypeOffline)

	// TODO: Automatically open browser or web view
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authUrl)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		return nil, fmt.Errorf("failed to read auth code: %v", err)
	}

	tok, err := cfg.Exchange(ctx, authCode)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange for auth token: %v", err)
	}

	return tok, nil
}

func (m *Monitor) Stop() {
	close(m.stop)
}

func (m *Monitor) poll(ctx context.Context) error {
	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error { return m.pollCalendarReminders(ctx) })
	g.Go(func() error { return m.pollEmails(ctx) })

	if err := g.Wait(); err != nil {
		return err
	}

	return nil
}

func (m *Monitor) pollCalendarReminders(ctx context.Context) error {
	slog.Debug("polling calendar reminders")

	events, err := m.calSvc.Events.List("primary").
		ShowDeleted(false).
		SingleEvents(true).
		TimeMin(time.Now().Format(time.RFC3339)).
		MaxResults(10).
		OrderBy("startTime").
		Context(ctx).
		Do()

	if err != nil {
		return fmt.Errorf("failed to get events: %v", err)
	}

	// TODO: Get reminders instead of events
	// TODO: Queue notification
	// TODO: Cache already fetched reminders
	for _, e := range events.Items {
		eventName := e.Summary
		eventDesc := e.Description
		eventStart, err := time.Parse(time.RFC3339, e.Start.DateTime)

		if err != nil {
			slog.Error("failed to parse event start", "event_name", eventName, "event_start", e.Start.DateTime)
			continue
		}

		m.cCalendarReminder <- &CalendarReminder{eventName, eventDesc, eventStart}
	}

	return nil
}

func (m *Monitor) pollEmails(ctx context.Context) error {
	slog.Debug("polling emails")

	gc, err := NewGmailClient(ctx, m.httpClient)
	if err != nil {
		return fmt.Errorf("failed to poll emails: %v", err)
	}

	gc.Sync(ctx)
	for _, e := range gc.cache.emails {
		m.cEmail <- e
	}

	return nil
}
