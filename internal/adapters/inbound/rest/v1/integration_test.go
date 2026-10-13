package rest_test

import (
	"bakery/cmd/bakery/wire"
	"bakery/cmd/bakery/wire/dep"
	"bakery/infrastructure/config"
	api "bakery/internal/adapters/inbound/rest/v1"
	"bakery/internal/domains/orders"
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Runs only against an explicitly supplied, disposable test DB. Never DATABASE_URL.
func TestCheckoutPostgres(t *testing.T) {
	u := os.Getenv("TEST_DATABASE_URL")
	if u == "" {
		t.Skip("set TEST_DATABASE_URL to a fresh disposable database ending in _test")
	}
	ctx := context.Background()
	cfg, e := pgxpool.ParseConfig(u)
	if e != nil {
		t.Fatal(e)
	}
	// Match the production pooler-safe mode, including JSON parameter encoding.
	cfg.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeExec
	db, e := pgxpool.NewWithConfig(ctx, cfg)
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	var name string
	if e = db.QueryRow(ctx, "SELECT current_database()").Scan(&name); e != nil {
		t.Fatal(e)
	}
	if !strings.HasSuffix(name, "_test") {
		t.Fatal("refusing non-test database")
	}
	// Fresh test DB only: CREATE SCHEMA fails rather than deleting existing tables.
	raw, e := os.ReadFile("../../../../../db/schema.sql")
	if e != nil {
		t.Fatal(e)
	}
	if _, e = db.Exec(ctx, string(raw), pgx.QueryExecModeSimpleProtocol); e != nil {
		t.Fatal(e)
	}
	var hasSettings bool
	if e = db.QueryRow(ctx, "SELECT to_regclass('app_schema.store_settings') IS NOT NULL").Scan(&hasSettings); e != nil {
		t.Fatal(e)
	}
	// dbmate dump may already include this migration's structure.
	if !hasSettings {
		raw, e = os.ReadFile("../../../../../db/migrations/20261010120000_api_settings_and_evidence.sql")
		if e != nil {
			t.Fatal(e)
		}
		up := strings.Split(string(raw), "-- migrate:down")[0]
		if _, e = db.Exec(ctx, up, pgx.QueryExecModeSimpleProtocol); e != nil {
			t.Fatal(e)
		}
	}
	// A schema-only dump does not contain the migration's store data.
	if _, e = db.Exec(ctx, `INSERT INTO app_schema.store_settings(key,value) VALUES('store','{"store_name":"TongBai","description":"Test store","delivery_fee_thb":50,"free_delivery_threshold_thb":800,"accepting_orders":true}') ON CONFLICT(key) DO NOTHING`); e != nil {
		t.Fatal(e)
	}
	const actor = "10000000-0000-4000-8000-000000000001"
	const cat = "20000000-0000-4000-8000-000000000001"
	const prod = "30000000-0000-4000-8000-000000000001"
	if _, e = db.Exec(ctx, "INSERT INTO app_schema.users(id,role) VALUES($1,'admin')", actor); e != nil {
		t.Fatal(e)
	}
	if _, e = db.Exec(ctx, "INSERT INTO app_schema.categories(id,slug,name_th,is_active) VALUES($1,'bread','ขนมปัง',true)", cat); e != nil {
		t.Fatal(e)
	}
	if _, e = db.Exec(ctx, "INSERT INTO app_schema.products(id,category_id,name,slug,description_th,price_satang,stock_quantity,is_active,is_featured) VALUES($1,$2,'Bread','bread','Fresh',10000,3,true,false)", prod, cat); e != nil {
		t.Fatal(e)
	}
	app := wire.NewApplication(dep.Dependencies{Db: db}, config.Config{})
	s := api.NewRestServer(app.NewRestUsecases(), api.Auth{Secret: strings.Repeat("s", 32), AdminToken: "review-admin", AdminID: actor}, "", "")
	h := s.Handler()
	call := func(method, path, body, token, key string, want int) *httptest.ResponseRecorder {
		t.Helper()
		r := httptest.NewRequest(method, path, strings.NewReader(body))
		if token != "" {
			r.Header.Set("Authorization", "Bearer "+token)
		}
		if key != "" {
			r.Header.Set("Idempotency-Key", key)
		}
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code != want {
			t.Fatalf("%s %s: %d want %d: %s", method, path, w.Code, want, w.Body.String())
		}
		return w
	}
	body := `{"customer":{"name":"Test Buyer","phone":"0812345678"},"fulfillment":{"type":"delivery","address":"Bangkok"},"items":[{"productId":"` + prod + `","quantity":2}],"paymentMethod":"bank_transfer"}`
	call("GET", "/products", "", "", "", 200)
	call("GET", "/categories", "", "", "", 200)
	var imageBody bytes.Buffer
	imageForm := multipart.NewWriter(&imageBody)
	imageFile, e := imageForm.CreateFormFile("image", "bread.png")
	if e != nil {
		t.Fatal(e)
	}
	_, _ = imageFile.Write([]byte("\x89PNG\r\n\x1a\n"))
	_ = imageForm.Close()
	imageRequest := httptest.NewRequest("POST", "/admin/products/"+prod+"/images", &imageBody)
	imageRequest.Header.Set("Authorization", "Bearer review-admin")
	imageRequest.Header.Set("Content-Type", imageForm.FormDataContentType())
	imageResponse := httptest.NewRecorder()
	h.ServeHTTP(imageResponse, imageRequest)
	if imageResponse.Code != 201 {
		t.Fatalf("upload image: %d %s", imageResponse.Code, imageResponse.Body.String())
	}
	var uploaded struct {
		Images []struct {
			PublicURL string `json:"public_url"`
		} `json:"images"`
	}
	if e = json.Unmarshal(imageResponse.Body.Bytes(), &uploaded); e != nil || len(uploaded.Images) != 1 {
		t.Fatalf("uploaded image metadata: %v %s", e, imageResponse.Body.String())
	}
	imageRead := call("GET", strings.TrimPrefix(uploaded.Images[0].PublicURL, "/api/v1"), "", "", "", 200)
	if imageRead.Header().Get("Content-Type") != "image/png" || !bytes.Equal(imageRead.Body.Bytes(), []byte("\x89PNG\r\n\x1a\n")) {
		t.Fatal("uploaded image content mismatch")
	}
	call("POST", "/checkout/quote", body, "", "", 200)
	var order orders.Data
	w := call("POST", "/orders", body, "", "review-order-0001", 201)
	if e = json.Unmarshal(w.Body.Bytes(), &order); e != nil {
		t.Fatal(e)
	}
	if order.Total != 250 || order.AccessToken == "" {
		t.Fatal(order)
	}
	var retry orders.Data
	w = call("POST", "/orders", body, "", "review-order-0001", 201)
	_ = json.Unmarshal(w.Body.Bytes(), &retry)
	if retry.ID != order.ID {
		t.Fatal("duplicate order created")
	}
	call("POST", "/orders", strings.Replace(body, `"quantity":2`, `"quantity":1`, 1), "", "review-order-0001", 409)
	call("POST", "/orders", body, "", "review-order-0002", 409)
	call("GET", "/orders/"+order.ID, "", "", "", 403)
	call("GET", "/orders/"+order.ID, "", order.AccessToken, "", 200)
	call("POST", "/payments/"+order.PaymentID+"/submit", "", order.AccessToken, "", 409)
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	f, e := mw.CreateFormFile("slip", "slip.png")
	if e != nil {
		t.Fatal(e)
	}
	_, _ = f.Write([]byte("\x89PNG\r\n\x1a\n" + strings.Repeat("0", 32)))
	_ = mw.Close()
	r := httptest.NewRequest("POST", "/payments/"+order.PaymentID+"/slip", &buf)
	r.Header.Set("Authorization", "Bearer "+order.AccessToken)
	r.Header.Set("Content-Type", mw.FormDataContentType())
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != 201 {
		t.Fatalf("slip %d %s", w.Code, w.Body.String())
	}
	call("POST", "/payments/"+order.PaymentID+"/submit", "", order.AccessToken, "", 200)
	call("POST", "/admin/payments/"+order.PaymentID+"/approve", "", "review-admin", "", 200)
	call("POST", "/admin/payments/"+order.PaymentID+"/approve", "", "review-admin", "", 200)
	var stock, count int
	if e = db.QueryRow(ctx, "SELECT stock_quantity FROM app_schema.products WHERE id=$1", prod).Scan(&stock); e != nil {
		t.Fatal(e)
	}
	if stock != 1 {
		t.Fatalf("double stock deduction: %d", stock)
	}
	if e = db.QueryRow(ctx, "SELECT count(*) FROM app_schema.notification_outbox").Scan(&count); e != nil {
		t.Fatal(e)
	}
	if count != 1 {
		t.Fatalf("outbox count %d", count)
	}
	call("PATCH", "/admin/orders/"+order.ID+"/status", `{"status":"completed"}`, "review-admin", "", 409)
	for _, path := range []string{"/admin/dashboard", "/admin/orders", "/admin/customers", "/admin/settings", "/admin/products"} {
		call("GET", path, "", "review-admin", "", 200)
	}
	call("PATCH", "/admin/products/"+prod, `{"price_thb":0,"is_featured":false}`, "review-admin", "", 200)
	// Two buyers competing for the final unit: only one may reserve it.
	small := strings.Replace(body, `"quantity":2`, `"quantity":1`, 1)
	results := make(chan int, 2)
	for _, key := range []string{"concurrent-order-001", "concurrent-order-002"} {
		go func(key string) {
			r := httptest.NewRequest("POST", "/orders", strings.NewReader(small))
			r.Header.Set("Idempotency-Key", key)
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)
			results <- w.Code
		}(key)
	}
	a, b := <-results, <-results
	counts := map[int]int{}
	counts[a]++
	counts[b]++
	if counts[201] != 1 || counts[409] != 1 {
		t.Fatalf("concurrent results %d %d", a, b)
	}
	call("PATCH", "/admin/products/"+prod, `{"stock_quantity":0}`, "review-admin", "", 409)
	// Force a later payment write to fail: customer/order/items/reservations must rollback.
	const prod2 = "30000000-0000-4000-8000-000000000002"
	if _, e = db.Exec(ctx, "INSERT INTO app_schema.products(id,category_id,name,slug,description_th,price_satang,stock_quantity,is_active,is_featured) VALUES($1,$2,'Cake','cake','Fresh',15000,5,true,false)", prod2, cat); e != nil {
		t.Fatal(e)
	}
	if _, e = db.Exec(ctx, `CREATE FUNCTION app_schema.fail_test_payment() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN IF NEW.idempotency_key='rollback-payment-001' THEN RAISE EXCEPTION 'forced test failure'; END IF; RETURN NEW; END $$; CREATE TRIGGER fail_test_payment BEFORE INSERT ON app_schema.payments FOR EACH ROW EXECUTE FUNCTION app_schema.fail_test_payment();`, pgx.QueryExecModeSimpleProtocol); e != nil {
		t.Fatal(e)
	}
	failed := strings.Replace(body, prod, prod2, 1)
	call("POST", "/orders", failed, "", "rollback-payment-001", 500)
	if e = db.QueryRow(ctx, "SELECT count(*) FROM app_schema.orders WHERE idempotency_key='rollback-payment-001'").Scan(&count); e != nil || count != 0 {
		t.Fatalf("order not rolled back: %d %v", count, e)
	}
	if e = db.QueryRow(ctx, "SELECT count(*) FROM app_schema.inventory_reservations WHERE product_id=$1", prod2).Scan(&count); e != nil || count != 0 {
		t.Fatalf("reservation not rolled back: %d %v", count, e)
	}
	// Cancellation releases reservations so another checkout can buy the same stock.
	var cancel orders.Data
	w = call("POST", "/orders", failed, "", "cancel-payment-0001", 201)
	_ = json.Unmarshal(w.Body.Bytes(), &cancel)
	call("PATCH", "/admin/orders/"+cancel.ID+"/status", `{"status":"cancelled"}`, "review-admin", "", 200)
	call("POST", "/payments/"+cancel.PaymentID+"/submit", "", cancel.AccessToken, "", 409)
	var review orders.Data
	w = call("POST", "/orders", failed, "", "after-cancel-00001", 201)
	if e = json.Unmarshal(w.Body.Bytes(), &review); e != nil {
		t.Fatal(e)
	}
	var evidence bytes.Buffer
	slipWriter := multipart.NewWriter(&evidence)
	slipFile, e := slipWriter.CreateFormFile("slip", "review.png")
	if e != nil {
		t.Fatal(e)
	}
	_, _ = slipFile.Write([]byte("\x89PNG\r\n\x1a\n" + strings.Repeat("0", 32)))
	_ = slipWriter.Close()
	req := httptest.NewRequest("POST", "/payments/"+review.PaymentID+"/slip", &evidence)
	req.Header.Set("Authorization", "Bearer "+review.AccessToken)
	req.Header.Set("Content-Type", slipWriter.FormDataContentType())
	w = httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != 201 {
		t.Fatalf("review slip %d: %s", w.Code, w.Body.String())
	}
	call("POST", "/payments/"+review.PaymentID+"/submit", "", review.AccessToken, "", 200)
	// The LAST repository write fails: all earlier stock/payment/history writes
	// must rollback, proving the composite passed one transaction to every repo.
	if _, e = db.Exec(ctx, `CREATE FUNCTION app_schema.fail_test_outbox() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'forced outbox failure'; END $$; CREATE TRIGGER fail_test_outbox BEFORE INSERT ON app_schema.notification_outbox FOR EACH ROW EXECUTE FUNCTION app_schema.fail_test_outbox();`, pgx.QueryExecModeSimpleProtocol); e != nil {
		t.Fatal(e)
	}
	call("POST", "/admin/payments/"+review.PaymentID+"/approve", "", "review-admin", "", 500)
	if e = db.QueryRow(ctx, "SELECT stock_quantity FROM app_schema.products WHERE id=$1", prod2).Scan(&stock); e != nil || stock != 5 {
		t.Fatalf("approval stock not rolled back: %d %v", stock, e)
	}
	var state, paymentState string
	if e = db.QueryRow(ctx, "SELECT o.status,p.status FROM app_schema.orders o JOIN app_schema.payments p ON p.order_id=o.id WHERE o.id=$1", review.ID).Scan(&state, &paymentState); e != nil || state != "pending_verification" || paymentState != "submitted" {
		t.Fatalf("approval state not rolled back: %s %s %v", state, paymentState, e)
	}
	if e = db.QueryRow(ctx, "SELECT count(*) FROM app_schema.inventory_movements WHERE reference_id=$1", review.ID).Scan(&count); e != nil || count != 0 {
		t.Fatalf("inventory movements not rolled back: %d %v", count, e)
	}
	if e = db.QueryRow(ctx, "SELECT count(*) FROM app_schema.order_status_history WHERE order_id=$1 AND to_status='confirmed'", review.ID).Scan(&count); e != nil || count != 0 {
		t.Fatalf("approval history not rolled back: %d %v", count, e)
	}
	if _, e = db.Exec(ctx, "DROP TRIGGER fail_test_outbox ON app_schema.notification_outbox; DROP FUNCTION app_schema.fail_test_outbox();", pgx.QueryExecModeSimpleProtocol); e != nil {
		t.Fatal(e)
	}
	// Simultaneous approvals must still deduct exactly once.
	for i := 0; i < 2; i++ {
		go func() {
			r := httptest.NewRequest("POST", "/admin/payments/"+review.PaymentID+"/approve", nil)
			r.Header.Set("Authorization", "Bearer review-admin")
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)
			results <- w.Code
		}()
	}
	if a, b := <-results, <-results; a != 200 || b != 200 {
		t.Fatalf("simultaneous approval: %d %d", a, b)
	}
	if e = db.QueryRow(ctx, "SELECT stock_quantity FROM app_schema.products WHERE id=$1", prod2).Scan(&stock); e != nil || stock != 3 {
		t.Fatalf("simultaneous stock deduction: %d %v", stock, e)
	}
}
