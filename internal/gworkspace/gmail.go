package gworkspace

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"slices"
	"sync"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"golang.org/x/sync/errgroup"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/googleapi"
)

const (
	credentialsFilePath = "credentials.json"
	tokenFilePath       = "token.json"
)

type HttpClient struct {
	*http.Client
}

func NewHttpClient() *HttpClient {
	return &HttpClient{
		Client: &http.Client{},
	}
}

func (c *HttpClient) Configure(ctx context.Context, scopes ...string) error {
	b, err := os.ReadFile(credentialsFilePath)
	if err != nil {
		return fmt.Errorf("error while reading credentials files (%s): %v", credentialsFilePath, err)
	}

	cfg, err := google.ConfigFromJSON(b, gmail.GmailReadonlyScope)
	if err != nil {
		return fmt.Errorf("error while configuring oauth: %v", err)
	}

	tok, err := getToken(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to get oauth token: %v", err)
	}

	c.Client = cfg.Client(ctx, tok)

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

type GmailHistoryId struct {
	id      uint64
	isValid bool
}

func NewGmailHistoryId() *GmailHistoryId {
	return &GmailHistoryId{
		id:      0,
		isValid: false,
	}
}

func NewGmailHistoryIdWithValue(id uint64) *GmailHistoryId {
	return &GmailHistoryId{
		id:      id,
		isValid: true,
	}
}

func (h *GmailHistoryId) SetId(id uint64) {
	h.id = id
	h.isValid = true
}

func (h *GmailHistoryId) GetId() uint64 {
	if !h.isValid {
		panic("attempted to get id of invalid GmailHistoryId. check IsValid() before accessing")
	}

	return h.id
}

func (h *GmailHistoryId) IsValid() bool {
	return h.isValid
}

func (h *GmailHistoryId) Clear() {
	h.isValid = false
}

type GmailMessage struct {
	To      string
	From    string
	Subject string
}

type GmailMonitor struct {
	mu  sync.Mutex
	svc *gmail.Service

	isInitialized bool
	historyId     *GmailHistoryId
	updateFreq    time.Duration

	msgsChan chan []*GmailMessage
}

func NewGmailMonitor(svc *gmail.Service, updateFreq time.Duration) *GmailMonitor {
	return &GmailMonitor{
		svc: svc,

		isInitialized: false,
		historyId:     NewGmailHistoryId(),
		updateFreq:    updateFreq,

		msgsChan: make(chan []*GmailMessage, 32),
	}
}

func (g *GmailMonitor) Initialize(ctx context.Context) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	err := g.refreshHistoryId(ctx)
	if err != nil {
		return fmt.Errorf("error while fetching latest history id: %v", err)
	}

	g.isInitialized = true

	return nil
}

func (g *GmailMonitor) Watch(ctx context.Context) error {
	ticker := time.NewTicker(g.updateFreq)

	tick := func() {
		slog.Debug("GmailMonitor Watch checking for new messages")

		err := g.CheckNow(ctx)
		if err != nil {
			slog.Error("error while checking for new messages", "error", err)
		}

		slog.Debug("GmailMonitor Watch waiting before checking again", "duration", g.updateFreq)
	}

	slog.Debug("starting GmailMonitor ticker")

	// tick once before waiting
	tick()

	for {
		select {
		case <-ticker.C:
			tick()
		case <-ctx.Done():
			return nil
		}
	}
}

func (g *GmailMonitor) CheckNow(ctx context.Context) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	slog.Debug("checking for new messages")

	msgs, err := g.fetchNewMessages(ctx)

	// 404 when history id is invalid
	if gerr, ok := err.(*googleapi.Error); ok && gerr.Code == http.StatusNotFound {
		slog.Debug("gmail responded 404 when fetching new messages. refreshing history id and trying again")

		err := g.refreshHistoryId(ctx)
		if err != nil {
			return fmt.Errorf("error while refreshing history id: %v", err)
		}

		msgs, err = g.fetchNewMessages(ctx)
		if err != nil {
			return fmt.Errorf("error while fetching new messages: %v", err)
		}
	}

	if err != nil {
		return fmt.Errorf("error while fetching new messages: %v", err)
	}

	slog.Info("received new messages from gmail", "numMessages", len(msgs))
	for _, msg := range msgs {
		slog.Debug("new messages", "to", msg.To, "from", msg.From, "subject", msg.Subject)
	}

	select {
	case g.msgsChan <- msgs:
	case <-ctx.Done():
	}

	return nil
}

func (g *GmailMonitor) Messages() <-chan []*GmailMessage {
	return g.msgsChan
}

func (g *GmailMonitor) fetchNewMessages(ctx context.Context) ([]*GmailMessage, error) {
	if !g.isInitialized || !g.historyId.IsValid() {
		panic("attempted to check for messages, but GmailMonitor was not initialized. call Initialize() first")
	}

	slog.Debug("fetching new messages from gmail")

	msgIds := make([]string, 0)

	forEachPage := func(res *gmail.ListHistoryResponse) error {
		if res.HistoryId > g.historyId.GetId() {
			slog.Debug("updating history id", "old", g.historyId.GetId(), "new", res.HistoryId)
			g.historyId.SetId(res.HistoryId)
		}

		for _, h := range res.History {
			for _, m := range h.MessagesAdded {
				msgIds = append(msgIds, m.Message.Id)
			}
		}

		return nil
	}

	err := g.svc.Users.History.List("me").
		StartHistoryId(g.historyId.GetId()).
		HistoryTypes("messageAdded").
		LabelId("INBOX").
		Pages(ctx, forEachPage)

	if err != nil {
		return []*GmailMessage{}, fmt.Errorf("error while fetching history from gmail (last history id = %d): %v", g.historyId.GetId(), err)
	}

	group, ctx := errgroup.WithContext(ctx)
	group.SetLimit(16) // TODO: Make configurable

	msgs := make([]*GmailMessage, len(msgIds))

	for i, id := range msgIds {
		group.Go(func() error {
			res, err := g.svc.Users.Messages.Get("me", id).
				Context(ctx).
				Format("metadata").
				MetadataHeaders("To", "From", "Subject").
				Do()

			if err != nil {
				return fmt.Errorf("error while fetching metadata for message (message id = %s): %v", id, err)
			}

			msg := &GmailMessage{}
			for _, h := range res.Payload.Headers {
				switch h.Name {
				case "To":
					msg.To = h.Value
				case "From":
					msg.From = h.Value
				case "Subject":
					msg.Subject = h.Value
				}
			}

			msgs[i] = msg

			return nil
		})
	}

	err = group.Wait()
	if err != nil {
		slog.Error("error whie getting message details", "error", err)
	}

	msgs = slices.DeleteFunc(msgs, func(msg *GmailMessage) bool {
		return msg == nil
	})

	return msgs, nil
}

func (g *GmailMonitor) refreshHistoryId(ctx context.Context) error {
	res, err := g.svc.Users.GetProfile("me").
		Context(ctx).
		Do()

	if err != nil {
		return fmt.Errorf("error getting profile from Gmail: %v", err)
	}

	g.historyId.SetId(res.HistoryId)

	return nil
}
