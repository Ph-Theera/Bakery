package product

import (
	"bakery/internal/shared"
	"context"
	"math"
	"strings"
)

type Image struct {
	PublicURL string `json:"public_url"`
	AltText   string `json:"alt_text"`
	Primary   bool   `json:"is_primary"`
}
type Data struct {
	ID           string   `json:"id"`
	CategoryID   string   `json:"category_id"`
	Name         string   `json:"name"`
	Slug         string   `json:"slug"`
	Description  string   `json:"description_th"`
	Price        float64  `json:"price_thb"`
	ComparePrice *float64 `json:"compare_at_price_thb"`
	Stock        int      `json:"stock_quantity"`
	Active       bool     `json:"is_active"`
	Featured     bool     `json:"is_featured"`
	Images       []Image  `json:"images"`
}
type Category struct {
	ID   string `json:"id"`
	Slug string `json:"slug"`
	Name string `json:"name_th"`
}
type Query struct {
	Category      string
	Search        string
	ActiveOnly    bool
	Limit, Offset int
}
type Repository interface {
	List(context.Context, Query) ([]Data, error)
	Get(context.Context, string, bool) (Data, error)
	Save(context.Context, string, Data) (Data, error)
	Delete(context.Context, string, string) error
	Categories(context.Context) ([]Category, error)
}
type Service struct{ Repo Repository }

func (s Service) List(ctx context.Context, q Query) ([]Data, error) {
	if q.Limit < 1 || q.Limit > 100 || q.Offset < 0 || q.Category != "" && !shared.UUID(q.Category) {
		return nil, shared.ErrInvalid
	}
	return s.Repo.List(ctx, q)
}
func (s Service) Get(ctx context.Context, id string, public bool) (Data, error) {
	if len(id) == 0 || len(id) > 220 {
		return Data{}, shared.ErrInvalid
	}
	return s.Repo.Get(ctx, id, public)
}
func (s Service) Save(ctx context.Context, id string, p Data) (Data, error) {
	p.Name = strings.TrimSpace(p.Name)
	p.Slug = strings.TrimSpace(p.Slug)
	if !shared.UUID(p.CategoryID) || p.Name == "" || len([]rune(p.Name)) > 200 || p.Slug == "" || len(p.Slug) > 220 || math.IsNaN(p.Price) || math.IsInf(p.Price, 0) || p.Price < 0 || p.Price > 1e9 || p.Stock < 0 || p.Stock > 1e9 || p.ComparePrice != nil && (*p.ComparePrice < p.Price || math.IsNaN(*p.ComparePrice) || math.IsInf(*p.ComparePrice, 0) || *p.ComparePrice > 1e9) || len(p.Images) > 0 {
		return Data{}, shared.ErrInvalid
	}
	if id != "" && !shared.UUID(id) {
		return Data{}, shared.ErrInvalid
	}
	return s.Repo.Save(ctx, id, p)
}
