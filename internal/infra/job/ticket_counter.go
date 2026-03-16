package job

import (
	"context"
	"time"

	"github.com/guncv/ticket-reservation-server/internal/infra/log"
	"github.com/guncv/ticket-reservation-server/internal/service/event"
)

type TicketCounterJob struct {
	eventService event.EventService
	log          log.Logger
	interval     time.Duration
	stopCh       chan struct{}
}

func NewTicketCounterJob(eventService event.EventService, log log.Logger, interval time.Duration) *TicketCounterJob {
	return &TicketCounterJob{
		eventService: eventService,
		log:          log,
		interval:     interval,
		stopCh:       make(chan struct{}),
	}
}

func (j *TicketCounterJob) Start(ctx context.Context) {
	go j.run(ctx)
}

func (j *TicketCounterJob) Stop() {
	close(j.stopCh)
}

func (j *TicketCounterJob) run(ctx context.Context) {
	ticker := time.NewTicker(j.interval)
	defer ticker.Stop()

	j.updateAvailableTickets(ctx)

	for {
		select {
		case <-ticker.C:
			j.updateAvailableTickets(ctx)
		case <-j.stopCh:
			j.log.Info(ctx, "Ticket counter job stopped")
			return
		case <-ctx.Done():
			j.log.Info(ctx, "Ticket counter job context cancelled")
			return
		}
	}
}

func (j *TicketCounterJob) updateAvailableTickets(ctx context.Context) {
	if err := j.eventService.SyncAvailableTickets(ctx); err != nil {
		j.log.Error(ctx, "Failed to sync available tickets", err)
		return
	}

	j.log.Debug(ctx, "Synced available tickets for all events")
}
