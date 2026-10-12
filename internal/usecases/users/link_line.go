package users

import (
	"bakery/internal/domains/users"
	portin "bakery/internal/ports/port_in/users"
	usersout "bakery/internal/ports/port_out/users"
	"bakery/internal/shared"
	"context"
	"strings"
)

type LinkLineArgs struct {
	UserRepository usersout.Repository
	Identity       usersout.Identity
}
type linkLineHandler struct{ args LinkLineArgs }

func NewLinkLineUsecase(args LinkLineArgs) portin.LinkLineUsecase {
	return &linkLineHandler{args: args}
}
func (h *linkLineHandler) Handle(ctx context.Context, cmd portin.LinkLineCommand) (users.Profile, error) {
	if strings.TrimSpace(cmd.Token) == "" {
		return users.Profile{}, shared.ErrInvalid
	}
	p, e := h.args.Identity.Verify(ctx, cmd.Token)
	if e != nil {
		return users.Profile{}, e
	}
	return h.args.UserRepository.Link(ctx, p)
}
