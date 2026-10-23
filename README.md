# GoCommerce

REST API cho cửa hàng điện thoại và laptop, hỗ trợ quản lý sản phẩm, giỏ hàng và xử lý đơn hàng COD.

## Tính năng

- **Tài khoản:** đăng ký, đăng nhập JWT, phân quyền quản trị và quản lý người dùng.
- **Catalog:** danh mục, sản phẩm theo SKU, thương hiệu, thông số kỹ thuật; tìm kiếm, lọc giá và phân trang.
- **Giỏ hàng:** quản lý số lượng sản phẩm theo tài khoản, hiển thị giá hiện tại.
- **Đơn hàng:** checkout COD, theo dõi trạng thái và hủy đơn; lưu snapshot thông tin sản phẩm và giá.
- **Tính nhất quán:** transaction và row-level locking bảo vệ tồn kho; idempotency key tránh tạo đơn trùng khi retry; optimistic version check tránh cập nhật sản phẩm bị lỗi thời.

## Công nghệ

| Thành phần | Công nghệ |
|---|---|
| Backend | Go, Gin |
| Database | PostgreSQL, pgx, sqlc, golang-migrate |
| Authentication | JWT, bcrypt |
| Logging | Zerolog, request correlation ID |
| Testing | Go Testing, PostgreSQL integration tests |
| Container & CI | Docker, Docker Compose, GitHub Actions |
| API documentation | OpenAPI 3.0 |

## Kiến trúc

Ứng dụng tổ chức theo các lớp **Handler → Service → Repository**. Dependency được khởi tạo bằng constructor tại `internal/app`; các thao tác checkout và hủy đơn sử dụng chung transaction cho thay đổi đơn hàng, giỏ hàng và tồn kho.

```text
cmd/api/                 Application entrypoint
internal/app/            Dependency wiring and server lifecycle
internal/handler/v1/     HTTP handlers
internal/service/v1/     Business logic
internal/repository/     Persistence contracts and transaction store
internal/db/queries/     SQL source for sqlc
internal/db/migrations/  Versioned database migrations
internal/middleware/     Authentication, rate limiting and request logging
tests/integration/       PostgreSQL integration tests
```

Chi tiết thiết kế và ERD: [Architecture](docs/architecture.md).

## Chạy ứng dụng

Yêu cầu Docker với Linux containers, Docker Compose và PowerShell. Thực hiện tại thư mục gốc của repository:

```powershell
./scripts/init-env.ps1
docker compose up --build --wait
```

Nếu đã có `.env`, dùng `./scripts/init-env.ps1 -CompleteExisting` để bổ sung secret còn thiếu hoặc thay secret quá ngắn; cấu hình database hiện tại được giữ nguyên. Với volume đã khởi tạo, `DB_USER`, `DB_PASSWORD` và `DB_NAME` phải khớp database đang sử dụng.

Compose chờ PostgreSQL sẵn sàng, chạy migration rồi khởi động API.

| Endpoint | URL local |
|---|---|
| API | http://localhost:8080/api/v1 |
| Liveness | http://localhost:8080/health/live |
| Readiness | http://localhost:8080/health/ready |
| OpenAPI | http://localhost:8080/openapi.json |

Import OpenAPI vào Postman hoặc Swagger Editor để xem các endpoint và request schema.

### Demo luồng mua hàng

```powershell
./scripts/demo.ps1
```

Script tạo dữ liệu điện thoại/laptop, đăng ký và đăng nhập tài khoản, thêm sản phẩm vào giỏ, đặt đơn, gửi lại yêu cầu với cùng idempotency key và hủy đơn để kiểm tra hoàn kho.

### Chạy Go trực tiếp

Yêu cầu Go 1.26.5 trở lên. Sau khi chuẩn bị `.env`:

```powershell
docker compose up -d --wait db
docker compose run --rm migrate
go run ./cmd/api
```

Các lệnh chạy tại repository root. Biến môi trường của process được ưu tiên hơn `.env`; `DB_PORT` là cổng kết nối native, còn `DB_PUBLISH_PORT` là cổng database được Compose công bố ra host.

## Kiểm thử

```powershell
go test ./...
go vet ./...
go build -o ./bin/api.exe ./cmd/api
```

Integration tests yêu cầu `TEST_DATABASE_URL` trỏ tới database riêng có hậu tố `_test`. Mỗi test sử dụng schema cô lập và dọn schema khi kết thúc. Khi chưa cấu hình biến này, integration tests được bỏ qua.

Bộ kiểm thử bao phủ luồng mua hàng, phân quyền, idempotent checkout, rollback và hủy đơn đồng thời. Kịch bản 100 khách cạnh tranh tồn kho 10 xác nhận đúng 10 đơn thành công và tồn kho cuối bằng 0.

Xem [Testing](docs/verification.md) để chạy integration tests. [CI workflow](.github/workflows/ci.yml) kiểm tra định dạng, vet, tests với race detector, build binary và Docker image.

## Quy tắc nghiệp vụ

- Một cửa hàng, một kho, tiền tệ VND và thanh toán COD.
- Giỏ hàng không giữ chỗ tồn kho. Checkout lấy giá và tồn kho từ PostgreSQL.
- `POST /api/v1/orders` yêu cầu `Idempotency-Key` là UUID. Gửi lại cùng key và thông tin giao hàng trả về đơn đã tạo; thay đổi thông tin với key cũ trả `409`.
- Đơn đi qua `placed → processing → shipped → completed`. Khách hàng được hủy đơn ở trạng thái `placed`.
- Cập nhật sản phẩm yêu cầu `version` hiện tại; phiên bản cũ trả `409`.
- Access token có thời hạn 15 phút. Role và trạng thái tài khoản được kiểm tra tại mỗi request. Bootstrap API key chỉ dùng cho thao tác quản trị từ server hoặc CLI.

## Tài liệu

- [Architecture & ERD](docs/architecture.md)
- [Development guide](docs/learning-guide.md)
- [Testing & verification](docs/verification.md)
- [Deployment: Nginx, HTTPS, backup & restore](docs/deployment.md)
