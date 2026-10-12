package catalog

import (
	"bakery/internal/shared"
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

type Stock struct {
	ID, Name string
	Price    int64
	Quantity int64
}

func Validate(p Data) (Data, error) {
	p.Name = strings.TrimSpace(p.Name)
	p.Slug = strings.TrimSpace(p.Slug)
	if !shared.UUID(p.CategoryID) || p.Name == "" || len([]rune(p.Name)) > 200 || p.Slug == "" || len(p.Slug) > 220 || math.IsNaN(p.Price) || math.IsInf(p.Price, 0) || p.Price < 0 || p.Price > 1e9 || p.Stock < 0 || p.Stock > 1e9 || p.ComparePrice != nil && (*p.ComparePrice < p.Price || math.IsNaN(*p.ComparePrice) || math.IsInf(*p.ComparePrice, 0) || *p.ComparePrice > 1e9) || len(p.Images) > 0 {
		return p, shared.ErrInvalid
	}
	return p, nil
}
func ValidateStock(stock int64, reserved int64) error {
	if stock < reserved {
		return shared.ErrConflict
	}
	return nil
}
