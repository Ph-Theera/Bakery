package users

import (
	"bakery/internal/domains/users"
	"bakery/internal/shared"
	"context"
)

type Repository interface {
	Customers(context.Context, int, int) ([]users.Customer, error)
	Link(context.Context, users.Profile) (users.Profile, error)
	UpsertCustomer(context.Context, shared.Transaction, string, string, string) (string, error)
}
type Identity interface {
	Verify(context.Context, string) (users.Profile, error)
}
