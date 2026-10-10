package product

import (
	"errors"
	"strings"
	"unicode/utf8"
)

var ErrInvalid = errors.New("invalid product")
var ErrNotFound = errors.New("product not found")
var ErrDuplicate = errors.New("sku already exists")

type Data struct {
	ID          int64
	SKU         string
	Name        string
	PriceSatang int64
	Stock       int64
}

type Input struct {
	SKU         string
	Name        string
	PriceSatang int64
	Stock       int64
}

func (i Input) Normalize() (Input, error) {
	i.SKU = strings.TrimSpace(i.SKU)
	i.Name = strings.TrimSpace(i.Name)
	if i.SKU == "" || i.Name == "" || utf8.RuneCountInString(i.SKU) > 64 || utf8.RuneCountInString(i.Name) > 200 || strings.ContainsRune(i.SKU, 0) || strings.ContainsRune(i.Name, 0) || i.PriceSatang < 0 || i.Stock < 0 {
		return Input{}, ErrInvalid
	}
	return i, nil
}
