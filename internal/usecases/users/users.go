package users

import "context"

type Customer struct {
	ID         string `json:"id"`
	Name       string `json:"display_name"`
	Phone      string `json:"phone"`
	OrderCount int    `json:"order_count"`
}
type Profile struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
}
type Repository interface {
	Customers(context.Context, int, int) ([]Customer, error)
	Link(context.Context, Profile) (Profile, error)
}
type Identity interface {
	Verify(context.Context, string) (Profile, error)
}
