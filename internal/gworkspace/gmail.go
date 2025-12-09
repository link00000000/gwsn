package gworkspace

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"golang.org/x/sync/errgroup"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
)

var ErrFullSyncRequired = errors.New("full sync required")
var ErrAuthorizationRequired = errors.New("authorization required")
var ErrRetryLater = errors.New("retry later")

type cachedEmail struct {
	id      string
	subject string
	to      string
	from    string
}

type gmailCache struct {
	emails    []*cachedEmail
	historyId uint64
}

type GmailClient struct {
	svc   *gmail.Service
	cache *gmailCache
}

func NewGmailClient(ctx context.Context, httpClient *http.Client) (*GmailClient, error) {
	svc, err := gmail.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("failed to create gmail service: %v", err)
	}

	c := &GmailClient{
		svc: svc,
		cache: &gmailCache{
			emails: make([]*cachedEmail, 0),
		},
	}
	return c, nil
}

func (c *GmailClient) Sync(ctx context.Context) error {
	slog.Debug("performing gmail sync")

	// if we have already synced, this will not be 0
	if c.cache.historyId != 0 {
		err := c.PartialSync(ctx)

		if err != nil {
			if err == ErrFullSyncRequired {
				slog.Warn("partial sync failed. falling back to performing a full sync", "error", err)
				err := c.FullSync(ctx)

				if err != nil {
					return fmt.Errorf("failed to perform full sync: %v", err)
				}

				return nil
			} else {
				return fmt.Errorf("partial sync failed: %v", err)
			}
		}
	} else {
		err := c.FullSync(ctx)

		if err != nil {
			return fmt.Errorf("failed to perform full sync: %v", err)
		}
	}

	return nil
}

func (c *GmailClient) FullSync(ctx context.Context) error {
	slog.Debug("performing gmail full sync")

	// reset the cache
	slog.Debug("email cache reset")
	c.cache.emails = c.cache.emails[:0]
	c.cache.historyId = 0

	// TODO: handle pagination
	slog.Debug("calling Gmail API messages.list")
	res, err := c.svc.Users.Messages.List("me").
		IncludeSpamTrash(false).
		Context(ctx).
		Do()

	if err != nil {
		if err, ok := err.(*googleapi.Error); ok {
			switch err.Code {

			// 400
			case http.StatusUnauthorized, http.StatusForbidden:
				return ErrAuthorizationRequired

			case http.StatusTooManyRequests:
				return ErrRetryLater

			// 500
			case http.StatusInternalServerError, http.StatusBadGateway,
				http.StatusServiceUnavailable, http.StatusGatewayTimeout:
				return ErrRetryLater
			}
		}

		return fmt.Errorf("failed to call Gmail API messages.list: %v", err)
	}
	slog.Debug("got response from Gmail API messages.list", "http_status", res.HTTPStatusCode, "num_messages", len(res.Messages))

	n := 16 // TODO: make configurable

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(n + 1)
	results := make(chan *gmail.Message, n)

	g.Go(func() error {
		for _, m := range res.Messages {
			g.Go(func() error {
				slog.Debug("calling Gmail API messages.get", "id", m.Id)

				res, err := c.svc.Users.Messages.Get("me", m.Id).
					Context(ctx).
					Do()

				if err != nil {
					if err, ok := err.(*googleapi.Error); ok {
						switch err.Code {

						// 400
						case http.StatusUnauthorized, http.StatusForbidden:
							return ErrAuthorizationRequired

						case http.StatusTooManyRequests:
							return ErrRetryLater

						// 500
						case http.StatusInternalServerError, http.StatusBadGateway,
							http.StatusServiceUnavailable, http.StatusGatewayTimeout:
							return ErrRetryLater
						}
					}

					slog.Debug("error returned when calling Gmail API messages.get", "id", m.Id, "error", err)
					return fmt.Errorf("failed to call Gmail API messages.get: %v", err)
				}

				slog.Debug("got response from Gmail API messages.get", "id", m.Id, "subject", res.Header.Get("Subject"), "to", res.Header.Get("To"), "from", res.Header.Get("From"))
				results <- res

				return nil
			})
		}

		return nil
	})

	if err := g.Wait(); err != nil {
		return fmt.Errorf("failed to get message details: %v", err)
	}

	for res := range results {
		c.cache.emails = append(c.cache.emails, &cachedEmail{
			id:      res.Id,
			subject: res.Header.Get("Subject"),
			to:      res.Header.Get("To"),
			from:    res.Header.Get("From"),
		})
	}
	close(results)

	if len(res.Messages) > 0 {
		c.cache.historyId = res.Messages[0].HistoryId
	}

	return nil
}

func (c *GmailClient) PartialSync(ctx context.Context) error {
	slog.Debug("performing gmail partial sync")

	// TODO: handle pagination
	res, err := c.svc.Users.History.List("me").
		StartHistoryId(c.cache.historyId).
		HistoryTypes("messageAdded").
		Context(ctx).
		Do()

	if err != nil {
		if err, ok := err.(*googleapi.Error); ok {
			switch err.Code {

			// 400
			case http.StatusUnauthorized, http.StatusForbidden:
				return ErrAuthorizationRequired

			case http.StatusNotFound:
				// If the historyId is outside the available range of
				// history records, a full sync is required.
				//
				// See https://developers.google.com/workspace/gmail/api/guides/sync#limitations
				return ErrFullSyncRequired

			case http.StatusTooManyRequests:
				return ErrRetryLater

			// 500
			case http.StatusInternalServerError, http.StatusBadGateway,
				http.StatusServiceUnavailable, http.StatusGatewayTimeout:
				return ErrRetryLater
			}
		}

		return fmt.Errorf("failed to call Gmail API history.list: %v", err)
	}

	n := 16 // TODO: make configurable

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(n + 1)
	results := make(chan *gmail.Message, n)

	g.Go(func() error {
		for _, h := range res.History {
			for _, m := range h.MessagesAdded {
				g.Go(func() error {
					res, err := c.svc.Users.Messages.Get("me", m.Message.Id).
						Context(ctx).
						Do()

					if err != nil {
						if err, ok := err.(*googleapi.Error); ok {
							switch err.Code {

							// 400
							case http.StatusUnauthorized, http.StatusForbidden:
								return ErrAuthorizationRequired

							case http.StatusTooManyRequests:
								return ErrRetryLater

							// 500
							case http.StatusInternalServerError, http.StatusBadGateway,
								http.StatusServiceUnavailable, http.StatusGatewayTimeout:
								return ErrRetryLater
							}
						}

						return fmt.Errorf("failed to call Gmail API messages.get: %v", err)
					}

					results <- res

					return nil
				})
			}
		}

		return nil
	})

	if err := g.Wait(); err != nil {
		return fmt.Errorf("failed to get message details: %v", err)
	}

	for res := range results {
		c.cache.emails = append(c.cache.emails, &cachedEmail{
			id:      res.Id,
			subject: res.Header.Get("subject"),
			to:      res.Header.Get("to"),
			from:    res.Header.Get("from"),
		})
	}
	close(results)

	// update cached history id for next partial sync
	c.cache.historyId = res.HistoryId

	return nil
}
