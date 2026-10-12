package users

import (
	"bakery/internal/domains/users"
	"context"
)

type ListCustomersQuery struct{ Limit, Offset int }

type ListCustomersUsecase interface {
	Handle(context.Context, ListCustomersQuery) ([]users.Customer, error)
}
