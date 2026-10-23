# Kiểm chứng MVP và nội dung CV

## Kết quả local ngày 23/10/2026

- `go test -count=1 ./...` với TEST_DATABASE_URL trên PostgreSQL 18: PASS, gồm unit và integration tests.
- `go vet ./...`: PASS.
- `gofmt -l cmd internal pkg tests`: không có file cần sửa.
- Docker Compose trên project/volume/ports test riêng: build API non-root thành công, migration job hoàn thành, database/API healthy.
- `scripts/demo.ps1` chạy trên Docker API: PASS, retry trả cùng order ID và cancellation hoàn stock 10.
- Race detector chưa chạy local (Windows hiện không có compiler C trong PATH). CI có lệnh `go test -race`, nhưng workflow chưa được push/chạy trên GitHub.
- Không sửa `.env`, không chạy migration vào database hiện tại của bạn; không thay đổi DNS hoặc deploy VPS.

## Bộ kiểm thử

- Config validation và URL-encoding credential database.
- Bcrypt, chuẩn hóa email và ép đăng ký về Member.
- Rate limiter chặn mọi request vượt giới hạn, không phụ thuộc quyết định ghi log.
- Log không lộ body/password/token/API key; recovery không lộ panic.
- JWT đúng/sai chữ ký, hết hạn, thiếu token và customer không được quản trị.
- OpenAPI có operation cho mọi route thực thi, trừ endpoint tự phục vụ spec.
- Graceful shutdown chờ request đang chạy hoàn thành.
- PostgreSQL HTTP integration: đăng ký/login, CRUD, filter/pagination, optimistic version, cart, checkout/replay, quyền sở hữu, status transitions, deactivate và soft delete.
- 100 khách checkout đồng thời với stock 10: đúng 10 đơn thành công, stock cuối 0.
- Hủy đồng thời: hoàn kho một lần.
- 10 request checkout cùng user/key chạy đồng thời: cùng một order ID, chỉ trừ kho một lần.
- Thay đổi role có hiệu lực ở request kế tiếp; body quá 1 MiB trả 413.
- Thiếu sản phẩm thứ hai: rollback stock sản phẩm thứ nhất, không tạo đơn, giữ nguyên giỏ.
- Migration Catalog/Shopping: up → down → up trên schema test riêng.

Chỉ `go test ./...` không có TEST_DATABASE_URL chưa chứng minh integration tests chạy. Không dùng kết quả concurrency này làm benchmark RPS/p95 vì chưa phải load test có đo hiệu năng chuẩn.

## Checklist trước khi gửi CV

- [ ] Tự chạy full stack và script demo trên máy bạn.
- [ ] Hiểu và giải thích được transaction/locking/idempotency.
- [ ] Review diff, xác nhận không có `.env` hoặc secret được Git theo dõi.
- [ ] Đưa README/OpenAPI/tests lên repository mà nhà tuyển dụng truy cập được.
- [ ] Push rồi xác nhận CI trên GitHub thực sự xanh.
- [ ] Nếu ghi deployed, xác nhận URL domain có HTTPS và hoạt động từ máy khác.
- [ ] Không ghi frontend/payment/email/Redis/RabbitMQ/CD/monitoring là đã hoàn thành.

## Nội dung CV có thể dùng sau khi tự kiểm chứng

**GoCommerce — Electronics Store REST API | Dự án cá nhân**

- Xây dựng REST API bằng Go/Gin cho tài khoản, catalog điện thoại/laptop, giỏ hàng và đơn COD; phân quyền quản trị bằng JWT và trạng thái user hiện tại.
- Sử dụng PostgreSQL, pgx, sqlc và SQL migrations; triển khai tìm kiếm/phân trang, soft delete và optimistic version check.
- Xử lý checkout/hủy đơn bằng transaction và row-level locking, có idempotency khi retry; integration test xác nhận 100 khách cạnh tranh tồn kho 10 chỉ tạo đúng 10 đơn.
- Viết unit/integration tests, Dockerfile nhiều stage và Docker Compose; chuẩn bị CI kiểm tra format, vet, test/race và build.

Chỉ thêm “triển khai trên Linux VPS với Nginx/HTTPS tại api.doanquanghuy.id.vn” sau khi deploy thật. Chỉ đổi “chuẩn bị CI” thành “tích hợp CI” sau khi workflow đã chạy trên repository.
