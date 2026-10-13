package product

import (
	"bakery/internal/domains/catalog"
	portin "bakery/internal/ports/port_in/catalog"
	portout "bakery/internal/ports/port_out/catalog"
	"bakery/internal/shared"
	"context"
	"strings"
)

type UploadProductImageArgs struct{ Images portout.ImageRepository }
type uploadProductImageHandler struct{ args UploadProductImageArgs }

func NewUploadProductImageUsecase(args UploadProductImageArgs) portin.UploadProductImageUsecase {
	return &uploadProductImageHandler{args: args}
}
func (h *uploadProductImageHandler) Handle(ctx context.Context, c portin.UploadProductImageCommand) (catalog.Data, error) {
	c.AltText = strings.TrimSpace(c.AltText)
	if !shared.UUID(c.ProductID) || len(c.Content) == 0 || len(c.Content) > 5<<20 || len([]rune(c.AltText)) > 255 || (c.Mime != "image/jpeg" && c.Mime != "image/png" && c.Mime != "image/webp") {
		return catalog.Data{}, shared.ErrInvalid
	}
	return h.args.Images.AddImage(ctx, c.ProductID, c.Mime, c.AltText, c.Content)
}

type GetProductImageArgs struct{ Images portout.ImageRepository }
type getProductImageHandler struct{ args GetProductImageArgs }

func NewGetProductImageUsecase(args GetProductImageArgs) portin.GetProductImageUsecase {
	return &getProductImageHandler{args: args}
}
func (h *getProductImageHandler) Handle(ctx context.Context, q portin.GetProductImageQuery) (portin.ProductImageContent, error) {
	if !shared.UUID(q.ID) {
		return portin.ProductImageContent{}, shared.ErrInvalid
	}
	mime, content, err := h.args.Images.GetImage(ctx, q.ID)
	return portin.ProductImageContent{Mime: mime, Content: content}, err
}
