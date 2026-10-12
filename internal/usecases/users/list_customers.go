package users

import (
	"bakery/internal/domains/users"
	portin "bakery/internal/ports/port_in/users"
	usersout "bakery/internal/ports/port_out/users"
	"bakery/internal/shared"
	"context"
)

type ListCustomersArgs struct{ UserRepository usersout.Repository }
type listCustomersHandler struct{ args ListCustomersArgs }

func NewListCustomersUsecase(args ListCustomersArgs) portin.ListCustomersUsecase {
	return &listCustomersHandler{args: args}
}
func (h *listCustomersHandler) Handle(ctx context.Context, q portin.ListCustomersQuery) ([]users.Customer, error) {
	if q.Limit < 1 || q.Limit > 100 || q.Offset < 0 {
		return nil, shared.ErrInvalid
	}
	return h.args.UserRepository.Customers(ctx, q.Limit, q.Offset)
}
