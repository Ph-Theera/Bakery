package users

import (
	"bakery/internal/domains/users"
	"context"
)

type LinkLineCommand struct{ Token string }

type LinkLineUsecase interface {
	Handle(context.Context, LinkLineCommand) (users.Profile, error)
}
