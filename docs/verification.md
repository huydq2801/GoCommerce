# Testing & verification

## Chạy kiểm thử

```powershell
go test ./...
go vet ./...
```

Unit tests chạy độc lập với PostgreSQL. Integration tests chỉ chạy khi có `TEST_DATABASE_URL`.

### PostgreSQL integration tests

Tạo container riêng cho kiểm thử:

```powershell
docker run -d --name gocommerce-test -p 127.0.0.1:55433:5432 -e POSTGRES_USER=gocommerce_test -e POSTGRES_PASSWORD=local-test-only -e POSTGRES_DB=gocommerce_test postgres:18-alpine
docker exec gocommerce-test pg_isready -U gocommerce_test -d gocommerce_test
```

Sau khi PostgreSQL báo `accepting connections`:

```powershell
$env:TEST_DATABASE_URL = 'postgresql://gocommerce_test:local-test-only@localhost:55433/gocommerce_test?sslmode=disable'
go test -v ./tests/integration
```

Kết thúc kiểm thử, dọn container test:

```powershell
docker rm -f gocommerce-test
```

Tests yêu cầu tên database kết thúc bằng `_test`, tạo schema riêng cho từng test và dọn schema khi hoàn tất. Không trỏ TEST_DATABASE_URL tới database ứng dụng. Các credential trong ví dụ chỉ dùng cho container test local.

Trên môi trường có compiler C, chạy race detector:

```sh
go test -race ./...
```

## Phạm vi kiểm thử

| Nhóm | Hành vi được kiểm tra |
|---|---|
| Configuration | Validate environment và URL-encode database credential |
| Registration | Chuẩn hóa email, hash password và gán Member mặc định |
| Authentication | Token hợp lệ, sai chữ ký, hết hạn; quyền Admin và trạng thái user |
| Middleware | Rate limit, giới hạn body, recovery và log không chứa dữ liệu nhạy cảm |
| API contract | OpenAPI bao phủ các route; validation và HTTP error mapping |
| Catalog | CRUD, lọc, phân trang, unique constraints và optimistic version |
| Checkout | Snapshot giá, idempotency và quyền sở hữu đơn |
| Transactions | Rollback tồn kho và giỏ hàng khi một sản phẩm không đủ hàng |
| Cancellation | Hủy lặp/đồng thời chỉ hoàn kho một lần |
| Lifecycle | Graceful shutdown chờ request đang xử lý |
| Migrations | Catalog/Shopping up → down → up trên schema test riêng |

## Kịch bản concurrency

- **100 khách, tồn kho 10:** đúng 10 đơn thành công, tồn kho cuối bằng 0.
- **10 request cùng checkout key:** cùng một order ID, chỉ trừ kho một lần.
- **Hủy đơn đồng thời:** tồn kho được khôi phục một lần.
- **Checkout nhiều sản phẩm, sản phẩm sau thiếu hàng:** rollback lần trừ kho trước đó, không tạo đơn và giữ nguyên giỏ.

Đây là kiểm thử tính đúng đắn; không phải benchmark throughput hoặc latency.

## Kết quả kiểm chứng local

Bản MVP đã vượt qua unit/integration tests với PostgreSQL 18, `go vet`, kiểm tra định dạng và Docker Compose smoke test. Script demo xác nhận retry trả cùng order ID và hủy đơn hoàn kho đúng.

GitHub Actions được cấu hình để chạy format, vet, race tests, build binary và Docker image. Trạng thái từng lần chạy được theo dõi tại [Actions](https://github.com/huydq2801/GoCommerce/actions).
