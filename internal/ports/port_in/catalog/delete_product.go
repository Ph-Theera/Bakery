package catalog

import (
	"context"
)

type DeleteProductCommand struct{ ID, Actor string }

type DeleteProductUsecase interface {
	Handle(context.Context, DeleteProductCommand) error
}
