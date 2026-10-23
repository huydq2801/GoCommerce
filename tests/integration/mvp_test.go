package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"

	"gocommerce/internal/app"
	"gocommerce/internal/config"
	"gocommerce/internal/db/sqlc"
	v1dto "gocommerce/internal/dto/v1"
	"gocommerce/internal/repository"
	v1service "gocommerce/internal/service/v1"
)

const adminKey = "integration-admin-key-only-32-characters"
const jwtSecret = "integration-jwt-secret-only-32-characters"

func fixture(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("Set TEST_DATABASE_URL to a dedicated *_test database to run PostgreSQL integration tests")
	}
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatal("invalid TEST_DATABASE_URL")
	}
	if !strings.HasSuffix(cfg.ConnConfig.Database, "_test") {
		t.Fatal("integration tests require a dedicated database whose name ends with _test")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	admin, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if err = admin.Ping(ctx); err != nil {
		admin.Close()
		t.Fatal("test PostgreSQL unavailable")
	}
	schema := "gc_it_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if _, err = admin.Exec(ctx, "CREATE SCHEMA "+schema); err != nil {
		admin.Close()
		t.Fatal(err)
	}
	cfg.ConnConfig.RuntimeParams["search_path"] = schema + ",public"
	cfg.MaxConns = 20
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		pool.Close()
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if _, err := admin.Exec(cleanupCtx, "DROP SCHEMA "+schema+" CASCADE"); err != nil {
			t.Errorf("clean test schema: %v", err)
		}
		admin.Close()
	})
	for _, version := range []string{"000001_users", "000002_catalog", "000003_shopping"} {
		runMigration(t, pool, version+".up.sql")
	}
	return pool
}
func runMigration(t *testing.T, pool *pgxpool.Pool, name string) {
	t.Helper()
	content, err := os.ReadFile(filepath.Join("..", "..", "internal", "db", "migrations", name))
	if err != nil {
		t.Fatal(err)
	}
	if _, err = pool.Exec(context.Background(), string(content)); err != nil {
		t.Fatalf("migration %s: %v", name, err)
	}
}
func handler(t *testing.T, pool *pgxpool.Pool) http.Handler {
	t.Helper()
	logger := zerolog.New(io.Discard)
	application, err := app.NewApplication(&config.Config{APIKey: adminKey, JWTSecret: jwtSecret, RatePerSecond: 10000, RateBurst: 10000}, pool, &logger)
	if err != nil {
		t.Fatal(err)
	}
	return application.Handler()
}
func request(t *testing.T, h http.Handler, method, path string, payload any, headers map[string]string, want int) map[string]any {
	t.Helper()
	var body io.Reader
	if payload != nil {
		encoded, err := json.Marshal(payload)
		if err != nil {
			t.Fatal(err)
		}
		body = strings.NewReader(string(encoded))
	}
	req := httptest.NewRequest(method, path, body)
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != want {
		t.Fatalf("%s %s: got %d want %d body=%s", method, path, w.Code, want, w.Body.String())
	}
	if w.Code == 204 {
		return nil
	}
	var result map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	return result
}
func registerAndLogin(t *testing.T, h http.Handler, email string) (string, string) {
	t.Helper()
	input := map[string]any{"name": "Demo Customer", "email": email, "age": 21, "password": "StrongPass1!"}
	created := request(t, h, "POST", "/api/v1/auth/register", input, nil, 201)["data"].(map[string]any)
	if created["level"] != "Member" {
		t.Fatal("registration did not default to Member")
	}
	login := request(t, h, "POST", "/api/v1/auth/login", map[string]any{"email": email, "password": "StrongPass1!"}, nil, 200)
	return created["uuid"].(string), login["data"].(map[string]any)["access_token"].(string)
}
func TestMVPHTTPFlow(t *testing.T) {
	pool := fixture(t)
	h := handler(t, pool)
	admin := map[string]string{"X-API-Key": adminKey}
	request(t, h, "GET", "/health/live", nil, nil, 200)
	request(t, h, "GET", "/health/ready", nil, nil, 200)
	request(t, h, "POST", "/api/v1/auth/register", map[string]any{"name": "Alex", "email": "x@example.com", "age": 21, "password": "StrongPass1!", "level": 1}, nil, 400)
	userUUID, token := registerAndLogin(t, h, "buyer@example.com")
	buyer := map[string]string{"Authorization": "Bearer " + token}
	request(t, h, "POST", "/api/v1/auth/login", map[string]any{"email": "buyer@example.com", "password": "wrong"}, nil, 401)
	request(t, h, "POST", "/api/v1/auth/register", map[string]any{"name": "Alex", "email": "buyer@example.com", "age": 21, "password": "StrongPass1!"}, nil, 409)
	request(t, h, "GET", "/api/v1/users/me", nil, buyer, 200)
	request(t, h, "GET", "/api/v1/users", nil, buyer, 403)
	request(t, h, "POST", "/api/v1/categories", map[string]any{"name": "Phones", "slug": "phones"}, buyer, 403)
	category := request(t, h, "POST", "/api/v1/categories", map[string]any{"name": "Phones", "slug": "phones"}, admin, 201)["data"].(map[string]any)
	categoryID := category["id"].(string)
	input := map[string]any{"category_id": categoryID, "sku": "PHONE-128", "name": "Demo Phone 128GB", "slug": "demo-phone-128", "brand": "Demo", "model": "Phone", "price_vnd": 15000000, "stock_quantity": 10, "specifications": map[string]string{"ram": "8GB", "storage": "128GB"}}
	product := request(t, h, "POST", "/api/v1/products", input, admin, 201)["data"].(map[string]any)
	productID := product["id"].(string)
	request(t, h, "POST", "/api/v1/products", input, admin, 409)
	request(t, h, "DELETE", "/api/v1/categories/"+categoryID, nil, admin, 409)
	filtered := request(t, h, "GET", "/api/v1/products?brand=demo&min_price=10000000&max_price=20000000&sort=price_asc&page=1&limit=1", nil, nil, 200)
	if len(filtered["data"].([]any)) != 1 || filtered["meta"].(map[string]any)["total"] != float64(1) {
		t.Fatal("incorrect filter/pagination")
	}
	request(t, h, "GET", "/api/v1/products?min_price=200&max_price=1", nil, nil, 400)
	request(t, h, "GET", "/api/v1/products?sort=arbitrary-sql", nil, nil, 400)
	input["version"] = 1
	input["is_active"] = true
	input["price_vnd"] = 14000000
	request(t, h, "PUT", "/api/v1/products/"+productID, input, admin, 200)
	request(t, h, "PUT", "/api/v1/products/"+productID, input, admin, 409)
	request(t, h, "PUT", "/api/v1/cart/items/"+productID, map[string]any{"quantity": 2}, buyer, 204)
	request(t, h, "PUT", "/api/v1/cart/items/"+productID, map[string]any{"quantity": 2}, buyer, 204)
	cart := request(t, h, "GET", "/api/v1/cart", nil, buyer, 200)["data"].(map[string]any)
	if cart["total_vnd"] != float64(28000000) {
		t.Fatal("cart quantity was added instead of replaced")
	}
	key := uuid.NewString()
	checkoutHeaders := map[string]string{"Authorization": "Bearer " + token, "Idempotency-Key": key}
	shipping := map[string]any{"recipient_name": "Alex", "phone": "0000000000", "address": "123 Example Street, Example City"}
	order := request(t, h, "POST", "/api/v1/orders", shipping, checkoutHeaders, 201)["data"].(map[string]any)
	orderID := order["id"].(string)
	if order["total_vnd"] != float64(28000000) {
		t.Fatal("checkout used incorrect price")
	}
	replay := request(t, h, "POST", "/api/v1/orders", shipping, checkoutHeaders, 200)["data"].(map[string]any)
	if replay["id"] != orderID {
		t.Fatal("retry created a different order")
	}
	shipping["address"] = "Different address"
	request(t, h, "POST", "/api/v1/orders", shipping, checkoutHeaders, 409)
	shipping["address"] = "123 Example Street, Example City"
	otherUUID, otherToken := registerAndLogin(t, h, "other@example.com")
	other := map[string]string{"Authorization": "Bearer " + otherToken}
	request(t, h, "GET", "/api/v1/orders/"+orderID, nil, other, 404)
	request(t, h, "POST", "/api/v1/orders/"+orderID+"/cancel", nil, other, 404)
	request(t, h, "PATCH", "/api/v1/users/"+userUUID, map[string]any{"level": 1}, buyer, 403)
	request(t, h, "PATCH", "/api/v1/users/"+otherUUID, map[string]any{"level": 1}, admin, 200)
	request(t, h, "GET", "/api/v1/users", nil, other, 200)
	request(t, h, "PATCH", "/api/v1/users/"+otherUUID, map[string]any{"level": 3}, admin, 200)
	request(t, h, "GET", "/api/v1/users", nil, other, 403)

	// Simultaneous cancellation must restore stock once, regardless of retries.
	q := sqlc.New(pool)
	u, err := q.GetUser(context.Background(), uuid.MustParse(userUUID))
	if err != nil {
		t.Fatal(err)
	}
	shop := v1service.NewShoppingService(repository.NewShoppingStore(pool))
	errs := make(chan error, 5)
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := shop.CancelOrder(context.Background(), u.UserID, uuid.MustParse(orderID))
			errs <- err
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	p, err := q.GetAdminProduct(context.Background(), uuid.MustParse(productID))
	if err != nil || p.StockQuantity != 10 {
		t.Fatalf("stock after concurrent cancellation=%d, err=%v", p.StockQuantity, err)
	}
	request(t, h, "PUT", "/api/v1/cart/items/"+productID, map[string]any{"quantity": 1}, buyer, 204)
	checkoutHeaders["Idempotency-Key"] = uuid.NewString()
	second := request(t, h, "POST", "/api/v1/orders", shipping, checkoutHeaders, 201)["data"].(map[string]any)["id"].(string)
	request(t, h, "PATCH", "/api/v1/admin/orders/"+second+"/status", map[string]any{"status": "completed"}, admin, 409)
	request(t, h, "PATCH", "/api/v1/admin/orders/"+second+"/status", map[string]any{"status": "processing"}, admin, 200)
	request(t, h, "POST", "/api/v1/orders/"+second+"/cancel", nil, buyer, 409)
	request(t, h, "PATCH", "/api/v1/admin/orders/"+second+"/status", map[string]any{"status": "shipped"}, admin, 200)
	request(t, h, "PATCH", "/api/v1/admin/orders/"+second+"/status", map[string]any{"status": "completed"}, admin, 200)
	request(t, h, "GET", "/api/v1/orders?page=1&limit=1", nil, buyer, 200)
	request(t, h, "GET", "/api/v1/users?search=buyer", nil, admin, 200)
	request(t, h, "PATCH", "/api/v1/users/"+userUUID, map[string]any{"name": "Updated Buyer"}, admin, 200)
	request(t, h, "DELETE", "/api/v1/products/"+productID, nil, admin, 204)
	request(t, h, "GET", "/api/v1/products/"+productID, nil, nil, 404)
	request(t, h, "DELETE", "/api/v1/users/"+userUUID, nil, admin, 204)
	request(t, h, "GET", "/api/v1/users/me", nil, buyer, 401)
}
func TestConcurrentCheckoutDoesNotOversell(t *testing.T) {
	pool := fixture(t)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	q := sqlc.New(pool)
	category, err := q.CreateCategory(ctx, sqlc.CreateCategoryParams{Name: "Laptops", Slug: "laptops"})
	if err != nil {
		t.Fatal(err)
	}
	product, err := q.CreateProduct(ctx, sqlc.CreateProductParams{CategoryID: category.ID, Sku: "LAST-LAPTOP", Name: "Last Laptop", Slug: "last-laptop", Brand: "Demo", Model: "Laptop", PriceVnd: 20000000, StockQuantity: 10, Specifications: json.RawMessage(`{}`)})
	if err != nil {
		t.Fatal(err)
	}
	service := v1service.NewShoppingService(repository.NewShoppingStore(pool))
	users := make([]int32, 100)
	for i := range users {
		user, err := q.CreateUser(ctx, sqlc.CreateUserParams{UserEmail: fmt.Sprintf("buyer%d@example.com", i), UserPassword: "not-used-for-login", UserFullname: "Concurrent buyer", UserStatus: 1, UserLevel: 3})
		if err != nil {
			t.Fatal(err)
		}
		users[i] = user.UserID
		if err = service.SetCartItem(ctx, user.UserID, product.ID, 1); err != nil {
			t.Fatal(err)
		}
	}
	start := make(chan struct{})
	outcomes := make(chan error, 100)
	var wg sync.WaitGroup
	for _, id := range users {
		wg.Add(1)
		go func(userID int32) {
			defer wg.Done()
			<-start
			_, _, err := service.Checkout(ctx, userID, uuid.New(), v1dto.CheckoutInput{RecipientName: "Buyer", Phone: "0000000000", Address: "123 Example Street, Example City"})
			outcomes <- err
		}(id)
	}
	close(start)
	wg.Wait()
	close(outcomes)
	success := 0
	for err := range outcomes {
		if err == nil {
			success++
		} else if !strings.Contains(strings.ToLower(err.Error()), "insufficient stock") {
			t.Fatalf("unexpected checkout failure: %v", err)
		}
	}
	current, err := q.GetAdminProduct(ctx, product.ID)
	if err != nil {
		t.Fatal(err)
	}
	count, err := q.CountOrders(ctx, 0)
	if err != nil {
		t.Fatal(err)
	}
	if success != 10 || current.StockQuantity != 0 || count != 10 {
		t.Fatalf("success=%d stock=%d orders=%d; want 10/0/10", success, current.StockQuantity, count)
	}
	t.Log("100 concurrent customers, stock=10: exactly 10 orders succeeded; final stock=0")
}
func TestCheckoutRollsBackEveryInventoryChange(t *testing.T) {
	pool := fixture(t)
	ctx := context.Background()
	q := sqlc.New(pool)
	category, err := q.CreateCategory(ctx, sqlc.CreateCategoryParams{Name: "Phones", Slug: "phones"})
	if err != nil {
		t.Fatal(err)
	}
	user, err := q.CreateUser(ctx, sqlc.CreateUserParams{UserEmail: "rollback@example.com", UserPassword: "not-used", UserFullname: "Buyer", UserStatus: 1, UserLevel: 3})
	if err != nil {
		t.Fatal(err)
	}
	ids := make([]uuid.UUID, 0, 2)
	for i := 0; i < 2; i++ {
		p, err := q.CreateProduct(ctx, sqlc.CreateProductParams{CategoryID: category.ID, Sku: fmt.Sprintf("PHONE-%d", i), Name: "Phone", Slug: fmt.Sprintf("phone-%d", i), Brand: "Demo", Model: "Phone", PriceVnd: 1000, StockQuantity: 1, Specifications: json.RawMessage(`{}`)})
		if err != nil {
			t.Fatal(err)
		}
		ids = append(ids, p.ID)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i].String() < ids[j].String() })
	if _, err = pool.Exec(ctx, "UPDATE products SET stock_quantity=0 WHERE id=$1", ids[1]); err != nil {
		t.Fatal(err)
	}
	shop := v1service.NewShoppingService(repository.NewShoppingStore(pool))
	for _, id := range ids {
		if err = shop.SetCartItem(ctx, user.UserID, id, 1); err != nil {
			t.Fatal(err)
		}
	}
	_, _, err = shop.Checkout(ctx, user.UserID, uuid.New(), v1dto.CheckoutInput{RecipientName: "Buyer", Phone: "0000000000", Address: "123 Example Street, Example City"})
	if err == nil {
		t.Fatal("checkout should fail on the second product")
	}
	first, err := q.GetAdminProduct(ctx, ids[0])
	if err != nil || first.StockQuantity != 1 {
		t.Fatal("earlier stock decrement was not rolled back")
	}
	orders, err := q.CountOrders(ctx, user.UserID)
	if err != nil || orders != 0 {
		t.Fatal("failed checkout created an order")
	}
	cart, err := shop.GetCart(ctx, user.UserID)
	if err != nil || len(cart.Items) != 2 {
		t.Fatal("failed checkout cleared cart")
	}
}
func TestNewMigrationsRoundTrip(t *testing.T) {
	pool := fixture(t)
	for _, name := range []string{"000003_shopping.down.sql", "000002_catalog.down.sql", "000002_catalog.up.sql", "000003_shopping.up.sql"} {
		runMigration(t, pool, name)
	}
	if _, err := sqlc.New(pool).ListCategories(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestSameCheckoutKeyConcurrentRequests(t *testing.T) {
	pool := fixture(t)
	ctx := context.Background()
	q := sqlc.New(pool)
	category, err := q.CreateCategory(ctx, sqlc.CreateCategoryParams{Name: "Phones", Slug: "phones"})
	if err != nil {
		t.Fatal(err)
	}
	product, err := q.CreateProduct(ctx, sqlc.CreateProductParams{CategoryID: category.ID, Sku: "RETRY-PHONE", Name: "Phone", Slug: "retry-phone", Brand: "Demo", Model: "Phone", PriceVnd: 10000, StockQuantity: 10, Specifications: json.RawMessage(`{}`)})
	if err != nil {
		t.Fatal(err)
	}
	user, err := q.CreateUser(ctx, sqlc.CreateUserParams{UserEmail: "retry@example.com", UserPassword: "not-used", UserFullname: "Buyer", UserStatus: 1, UserLevel: 3})
	if err != nil {
		t.Fatal(err)
	}
	service := v1service.NewShoppingService(repository.NewShoppingStore(pool))
	if err = service.SetCartItem(ctx, user.UserID, product.ID, 1); err != nil {
		t.Fatal(err)
	}
	key := uuid.New()
	type outcome struct {
		id  uuid.UUID
		err error
	}
	outcomes := make(chan outcome, 10)
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			order, _, err := service.Checkout(ctx, user.UserID, key, v1dto.CheckoutInput{RecipientName: "Buyer", Phone: "0000000000", Address: "123 Example Street, Example City"})
			outcomes <- outcome{id: order.ID, err: err}
		}()
	}
	wg.Wait()
	close(outcomes)
	id := uuid.Nil
	for result := range outcomes {
		if result.err != nil {
			t.Fatal(result.err)
		}
		if id == uuid.Nil {
			id = result.id
		}
		if id != result.id {
			t.Fatal("same key created different orders")
		}
	}
	current, err := q.GetAdminProduct(ctx, product.ID)
	if err != nil {
		t.Fatal(err)
	}
	count, err := q.CountOrders(ctx, user.UserID)
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 || current.StockQuantity != 9 {
		t.Fatalf("same-key concurrent retry: orders=%d stock=%d", count, current.StockQuantity)
	}
}
