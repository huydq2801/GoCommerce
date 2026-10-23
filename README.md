# GoCommerce — Electronics Store API

Backend MVP cho cửa hàng điện thoại và laptop, dùng để học Golang và trình bày portfolio Backend Intern.

**Trạng thái:** có luồng đăng ký → đăng nhập → xem sản phẩm → giỏ hàng → đặt đơn COD → hủy đơn hoặc quản trị giao hàng. Đây là REST API, chưa có giao diện cửa hàng và chưa triển khai public.

## Công nghệ

Go 1.26.5, Gin, PostgreSQL 18, pgx, sqlc, golang-migrate, JWT, bcrypt, Zerolog, Docker Compose. Giữ kiến trúc Handler → Service → Repository và constructor injection từ project đang học.

## Tính năng

- Đăng ký mặc định Member; đăng nhập cấp access token 15 phút; `/users/me`.
- Quản trị User CRUD, tìm kiếm/phân trang và soft delete.
- Category CRUD; sản phẩm theo SKU, brand/model, giá VND, stock và specifications JSON.
- Public search/filter/pagination; ngừng bán thay vì xóa lịch sử sản phẩm.
- Optimistic version check khi quản trị cập nhật sản phẩm; stock thay đổi cũng tăng version.
- Giỏ hàng theo user; đặt số lượng tuyệt đối, tối đa 50 SKU/giỏ và 1.000 đơn vị/SKU.
- Checkout trong transaction với row locks, snapshot giá và Idempotency-Key.
- Đơn COD: `placed → processing → shipped → completed`; customer chỉ hủy ở `placed`.
- JWT + kiểm tra user hiện tại trong PostgreSQL; admin role mới được quản trị. Bootstrap API key chỉ dùng từ CLI/server.
- Request ID, log metadata không chứa body/token/password; rate limiter trong process; health/readiness và graceful shutdown.
- Unit/regression tests và PostgreSQL integration tests; Dockerfile nhiều stage chạy non-root; workflow CI.

## Chạy nhanh bằng Docker — database mới

Yêu cầu Docker Desktop Linux engine, Docker Compose và PowerShell. Chạy từ **repository root**:

```powershell
./scripts/init-env.ps1
docker compose up --build --wait
Invoke-RestMethod http://localhost:8080/health/ready
./scripts/demo.ps1
```

Script init tạo `.env` với 3 giá trị ngẫu nhiên độc lập, không in secret. Script demo tạo category điện thoại/laptop, 2 sản phẩm, đăng ký/login, thêm giỏ, đặt đơn, retry cùng key và hủy đơn để kiểm tra hoàn kho. Mỗi lần chạy dùng mã riêng; không xóa dữ liệu cũ.

Nếu máy đã có database/volume hoặc `.env`:

1. Giữ nguyên DB_USER/DB_PASSWORD/DB_NAME đang dùng; thay đổi biến môi trường không đổi credential của một volume đã khởi tạo.
2. Chạy `./scripts/init-env.ps1 -CompleteExisting` để bổ sung JWT_SECRET hoặc thay API_KEY quá ngắn. Script tạo backup `.env.backup.*`, không thay credential database hiện có.
3. Kiểm tra DB_NAME và DB_USER trùng database thật. Nếu cổng 5432 đang bận, dùng DB_PUBLISH_PORT khác.
4. Backup database trước khi chạy migration trên dữ liệu cần giữ. Không dùng `docker compose down -v` để chữa lỗi kết nối.
5. `docker compose up --build --wait` sẽ chạy các migration chưa áp dụng trước khi khởi động API. Migration 000001 được giữ nguyên; 000002/000003 bổ sung Catalog/Shopping.

API: `http://localhost:8080`; OpenAPI: `http://localhost:8080/openapi.json`. Import URL/file `internal/apidocs/openapi.json` vào Postman hoặc Swagger Editor để xem tài liệu.

## Chạy Go trực tiếp

Cần Go 1.26.5+, PostgreSQL và migrate CLI. Chạy từ repository root:

```powershell
go mod download
# Sau khi chuẩn bị .env, PostgreSQL và áp dụng migration:
go run ./cmd/api
```

Nếu dùng Compose để chuẩn bị database:

```powershell
docker compose up -d --wait db
docker compose run --rm migrate
go run ./cmd/api
```

`.env` được đọc duy nhất tại entrypoint; environment của process có ưu tiên cao hơn file. DB_PORT là cổng Go kết nối khi chạy native; DB_PUBLISH_PORT là cổng Compose công bố ra host. Khi chạy cả Go native và Compose API, tránh để hai server dùng cùng SERVER_PORT.

## Kiểm thử

```powershell
go test ./...
go vet ./...
go build -o ./bin/api.exe ./cmd/api
```

Integration tests **skip** nếu chưa đặt TEST_DATABASE_URL. Để chạy thật, tạo database/container riêng:

```powershell
docker run -d --name gocommerce-test -p 127.0.0.1:55433:5432 -e POSTGRES_USER=gocommerce_test -e POSTGRES_PASSWORD=local-test-only -e POSTGRES_DB=gocommerce_test postgres:18-alpine
# Chờ pg_isready báo accepting connections:
docker exec gocommerce-test pg_isready -U gocommerce_test -d gocommerce_test
$env:TEST_DATABASE_URL = 'postgresql://gocommerce_test:local-test-only@localhost:55433/gocommerce_test?sslmode=disable'
go test -v ./tests/integration
# Sau khi xong, chỉ dọn container test do bạn vừa tạo:
docker rm -f gocommerce-test
```

Tests từ chối database không có hậu tố `_test`, tạo schema riêng mỗi test và dọn schema đó khi kết thúc. Không dùng database ứng dụng thật. Trên Linux có compiler C, chạy thêm `go test -race ./...`; CI cũng chạy race detector và integration tests.

## API chính

| Quyền | Endpoints |
|---|---|
| Public | `POST /api/v1/auth/register`, `POST /api/v1/auth/login` |
| Public | `GET /api/v1/categories`, `GET /api/v1/products`, `GET /api/v1/products/:uuid` |
| User JWT | `GET /api/v1/users/me`, `GET /api/v1/cart` |
| User JWT | `PUT/DELETE /api/v1/cart/items/:uuid`, `DELETE /api/v1/cart/items` |
| User JWT | `POST/GET /api/v1/orders`, `GET /api/v1/orders/:uuid`, `POST /api/v1/orders/:uuid/cancel` |
| Admin JWT hoặc bootstrap key | User CRUD; category/product writes; `GET /api/v1/admin/products/:uuid` |
| Admin JWT hoặc bootstrap key | `GET /api/v1/admin/orders`, `GET /api/v1/admin/orders/:uuid`, `PATCH /api/v1/admin/orders/:uuid/status` |

Public `POST /api/v1/users` là alias đăng ký giữ tương thích project cũ. User update dùng PATCH; product/category update dùng PUT. Không tự đưa API key vào frontend. Quản trị có thể dùng key từ CLI để cập nhật một user thành level 1, rồi đăng nhập bằng tài khoản đó. Schema giữ role 1 Admin, 2 Moderator, 3 Member; MVP chỉ Admin được quản trị.

## Quy tắc cần biết

- Một cửa hàng, một kho, một loại tiền VND; mỗi cấu hình điện thoại/laptop là SKU riêng.
- Giỏ hàng không giữ chỗ tồn kho. Giá trong giỏ là giá hiện tại, checkout tính lại tại database.
- Checkout cần header `Idempotency-Key` là UUID. Retry cùng key + địa chỉ đã chuẩn hóa trả đơn cũ; đổi input với key cũ trả 409. Key định danh một lần checkout, kể cả sau đó giỏ đã thay đổi.
- Lỗi thiếu kho/giỏ rỗng/trạng thái sai/version cũ: 409; không tìm thấy hoặc truy cập đơn người khác: 404.
- Khóa sản phẩm theo UUID tăng dần. Trừ kho, tạo đơn/chi tiết và xóa giỏ cùng commit hoặc cùng rollback.
- Hủy đơn lặp/concurrent hoàn kho một lần. Không hủy sau khi admin chuyển sang processing.
- Không có refresh/logout server-side ở MVP. Client xóa token khi đăng xuất; token bị đánh cắp có thể dùng tới hết hạn 15 phút. User bị khóa/xóa bị từ chối ngay ở request tiếp theo. Đổi password hiện chưa thu hồi token đang còn hạn.
- Chưa có Redis, queue, email, online payment, frontend, CD hay monitoring dashboard. Không ghi các phần đó là đã triển khai trên CV.

## Tài liệu

- [Kiến trúc, ERD và quyết định kỹ thuật](docs/architecture.md)
- [Đối chiếu khóa học và hướng dẫn đọc code](docs/learning-guide.md)
- [Domain/VPS/Nginx/HTTPS](docs/deployment.md)
- [Kết quả kiểm chứng và checklist CV](docs/verification.md)

CI nằm trong `.github/workflows/ci.yml`: format, vet, test/race, build binary và Docker image. Workflow chỉ được chuẩn bị ở local; trạng thái chạy trên GitHub cần kiểm tra sau khi bạn push.
