package product

import (
	"bakery/internal/domains/catalog"
	portin "bakery/internal/ports/port_in/catalog"
	"bakery/internal/shared"
	"context"
	"errors"
	"testing"
)

const productID = "11111111-1111-4111-8111-111111111111"

type imageRepoStub struct {
	called bool
	mime   string
	data   []byte
}

func (r *imageRepoStub) AddImage(_ context.Context, id, mime, alt string, data []byte) (catalog.Data, error) {
	r.called = true
	if id != productID || alt != "Bread" {
		return catalog.Data{}, errors.New("wrong input")
	}
	return catalog.Data{ID: id}, nil
}
func (r *imageRepoStub) GetImage(_ context.Context, id string) (string, []byte, error) {
	if id != productID {
		return "", nil, shared.ErrNotFound
	}
	return r.mime, r.data, nil
}

func TestUploadProductImage(t *testing.T) {
	for _, mime := range []string{"image/jpeg", "image/png", "image/webp"} {
		repo := &imageRepoStub{}
		got, err := NewUploadProductImageUsecase(UploadProductImageArgs{Images: repo}).Handle(context.Background(), portin.UploadProductImageCommand{ProductID: productID, Mime: mime, AltText: " Bread ", Content: []byte("bytes")})
		if err != nil || got.ID != productID || !repo.called {
			t.Fatalf("mime %s: product=%+v, err=%v", mime, got, err)
		}
	}
	for _, cmd := range []portin.UploadProductImageCommand{
		{ProductID: "bad", Mime: "image/png", Content: []byte("x")},
		{ProductID: productID, Mime: "image/svg+xml", Content: []byte("x")},
		{ProductID: productID, Mime: "image/png"},
		{ProductID: productID, Mime: "image/png", Content: make([]byte, (5<<20)+1)},
	} {
		repo := &imageRepoStub{}
		_, err := NewUploadProductImageUsecase(UploadProductImageArgs{Images: repo}).Handle(context.Background(), cmd)
		if !errors.Is(err, shared.ErrInvalid) || repo.called {
			t.Fatalf("expected validation error, got %v", err)
		}
	}
}

func TestGetProductImage(t *testing.T) {
	repo := &imageRepoStub{mime: "image/png", data: []byte("png")}
	got, err := NewGetProductImageUsecase(GetProductImageArgs{Images: repo}).Handle(context.Background(), portin.GetProductImageQuery{ID: productID})
	if err != nil || got.Mime != repo.mime || string(got.Content) != "png" {
		t.Fatalf("image=%+v, err=%v", got, err)
	}
}
