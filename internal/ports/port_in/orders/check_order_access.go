package orders

import (
	"context"
)

type CheckOrderAccessQuery struct{ ID, UserID string }

type CheckOrderAccessUsecase interface {
	Handle(context.Context, CheckOrderAccessQuery) (bool, error)
}
