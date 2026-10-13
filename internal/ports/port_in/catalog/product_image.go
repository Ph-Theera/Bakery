package catalog

import (
	"bakery/internal/domains/catalog"
	"context"
)

type UploadProductImageCommand struct {
	ProductID string
	Mime      string
	AltText   string
	Content   []byte
}

type UploadProductImageUsecase interface {
	Handle(context.Context, UploadProductImageCommand) (catalog.Data, error)
}

type GetProductImageQuery struct{ ID string }
type ProductImageContent struct {
	Mime    string
	Content []byte
}
type GetProductImageUsecase interface {
	Handle(context.Context, GetProductImageQuery) (ProductImageContent, error)
}
