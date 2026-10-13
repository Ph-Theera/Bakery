package catalog

import (
	"bakery/internal/domains/catalog"
	"context"
)

type ImageRepository interface {
	AddImage(context.Context, string, string, string, []byte) (catalog.Data, error)
	GetImage(context.Context, string) (string, []byte, error)
}
