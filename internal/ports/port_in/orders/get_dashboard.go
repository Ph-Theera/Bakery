package orders

import (
	"context"
)

type GetDashboardQuery struct{}

type GetDashboardUsecase interface {
	Handle(context.Context, GetDashboardQuery) (map[string]any, error)
}
