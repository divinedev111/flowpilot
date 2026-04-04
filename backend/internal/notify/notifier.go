package notify

import (
	"context"
	"log"
	"sync"

	"flowpilot/internal/models"
)

type Notifier interface {
	Name() string
	Send(ctx context.Context, event models.AlertEvent) error
}

type Dispatcher struct {
	notifiers []Notifier
}

func NewDispatcher(notifiers ...Notifier) *Dispatcher {
	return &Dispatcher{notifiers: notifiers}
}

// Dispatch sends an alert event to all configured notifiers concurrently
func (d *Dispatcher) Dispatch(ctx context.Context, events []models.AlertEvent) {
	var wg sync.WaitGroup
	for _, event := range events {
		for _, n := range d.notifiers {
			wg.Add(1)
			go func(n Notifier, e models.AlertEvent) {
				defer wg.Done()
				if err := n.Send(ctx, e); err != nil {
					log.Printf("notify %s error: %v", n.Name(), err)
				}
			}(n, event)
		}
	}
	wg.Wait()
}
