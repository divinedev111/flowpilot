package connectors

import (
	"context"

	"flowpilot/internal/models"
)

type Connector interface {
	Name() string
	FetchPositions(ctx context.Context) ([]models.Position, error)
}
