package googlecalendar

import (
	"context"

	"github.com/link00000000/gwsn/internal/services"
)

type googleCalendarService struct{}

var _ services.GoogleCalendarService = (*googleCalendarService)(nil)

func NewService() *googleCalendarService {
	return &googleCalendarService{}
}

// implements [services.GoogleCalendarService]
func (*googleCalendarService) Run(ctx context.Context) error {
	return nil
}
