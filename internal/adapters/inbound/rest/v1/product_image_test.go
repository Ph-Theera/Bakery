package rest

import (
	"bakery/internal/domains/catalog"
	portin "bakery/internal/ports/port_in/catalog"
	"bytes"
	"context"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
)

const imageTestID = "11111111-1111-4111-8111-111111111111"

type uploadImageStub struct{ called bool }

func (s *uploadImageStub) Handle(_ context.Context, c portin.UploadProductImageCommand) (catalog.Data, error) {
	s.called = true
	return catalog.Data{ID: c.ProductID}, nil
}

type getImageStub struct{}

func (getImageStub) Handle(context.Context, portin.GetProductImageQuery) (portin.ProductImageContent, error) {
	return portin.ProductImageContent{Mime: "image/png", Content: []byte("image")}, nil
}

func TestProductImageRoutes(t *testing.T) {
	upload := &uploadImageStub{}
	handler := NewRestServer(&Usecases{UploadProductImage: upload, GetProductImage: getImageStub{}}, Auth{AdminToken: "secret", AdminID: imageTestID}, "", "").Handler()
	var body bytes.Buffer
	form := multipart.NewWriter(&body)
	file, err := form.CreateFormFile("image", "sample.png")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = file.Write([]byte("\x89PNG\r\n\x1a\n"))
	_ = form.Close()
	makeRequest := func(token string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(http.MethodPost, "/admin/products/"+imageTestID+"/images", bytes.NewReader(body.Bytes()))
		r.Header.Set("Content-Type", form.FormDataContentType())
		if token != "" {
			r.Header.Set("Authorization", "Bearer "+token)
		}
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		return w
	}
	if got := makeRequest(""); got.Code != 403 || upload.called {
		t.Fatalf("unauthorized upload: %d", got.Code)
	}
	if got := makeRequest("secret"); got.Code != 201 || !upload.called {
		t.Fatalf("authorized upload: %d %s", got.Code, got.Body.String())
	}
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/products/images/"+imageTestID, nil))
	if w.Code != 200 || w.Header().Get("Content-Type") != "image/png" || w.Body.String() != "image" {
		t.Fatalf("image response: %d %s", w.Code, w.Body.String())
	}
}
