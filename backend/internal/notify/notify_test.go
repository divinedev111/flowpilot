package notify

import (
	"context"
	"sync/atomic"
	"testing"

	"flowpilot/internal/models"

	"github.com/stretchr/testify/assert"
)

type mockNotifier struct {
	name  string
	count atomic.Int32
}

func (m *mockNotifier) Name() string { return m.name }
func (m *mockNotifier) Send(ctx context.Context, event models.AlertEvent) error {
	m.count.Add(1)
	return nil
}

func TestDispatcher_SendsToAllNotifiers(t *testing.T) {
	n1 := &mockNotifier{name: "mock1"}
	n2 := &mockNotifier{name: "mock2"}

	d := NewDispatcher(n1, n2)

	events := []models.AlertEvent{
		{Message: "test alert 1", Severity: "high"},
		{Message: "test alert 2", Severity: "medium"},
	}

	d.Dispatch(context.Background(), events)

	assert.Equal(t, int32(2), n1.count.Load())
	assert.Equal(t, int32(2), n2.count.Load())
}

func TestDispatcher_NoNotifiers(t *testing.T) {
	d := NewDispatcher()
	d.Dispatch(context.Background(), []models.AlertEvent{{Message: "test"}})
	// Should not panic
}
