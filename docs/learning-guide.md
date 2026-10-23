# Học và giải thích được MVP

File khóa học bạn cung cấp là danh sách tiêu đề bài giảng, không phải nội dung/video đầy đủ. Việc đối chiếu dưới đây dựa trên tiêu đề và code project hiện tại.

| Nhóm bài | Áp dụng trong project |
|---|---|
| 01–21: HTTP, Gin, validation | Router, status codes, DTO và decode JSON; upload chưa cần cho MVP |
| 22–28, 57–59: middleware/logging | Request ID, recovery, rate limit; chỉ log metadata thay vì ghi nguyên password/body |
| 29–41, 60–70: CRUD, DTO, pagination | User CRUD thật, conflict/404, soft delete, search/pagination |
| 42–53: PostgreSQL, migration, pgx/sqlc | Schema constraints, pool, SQL queries, generated code |
| 54–56: shopping structure và shutdown | Module wiring; đóng HTTP trước, pool sau |
| 74–82: authentication | Bcrypt, JWT access token, role/status; refresh token/blacklist để sau |
| 83–86: Docker/build | Multi-stage Dockerfile, non-root runtime, Compose health + migration order |
| 71–73, 87–96: cache/email/queue/log stack | Hoãn Redis, email, RabbitMQ, Loki; CORS thêm khi có frontend khác origin |

## Thứ tự đọc code

1. `cmd/api/main.go`: config, pool, signal và cleanup.
2. `internal/app/app.go`, `user_module.go`, `catalog_module.go`: nối dependency.
3. User handler → service → repository → `internal/db/queries/users.sql`: một luồng CRUD dễ nắm.
4. Catalog DTO/filter → SQL: tiền VND, specifications và sort allowlist.
5. `internal/middleware/auth.go`: validate JWT rồi kiểm tra user hiện tại.
6. `internal/service/v1/shopping_service.go`: transaction checkout/cancel là phần quan trọng nhất.
7. `tests/integration/mvp_test.go`: hành vi hệ thống phải giữ đúng.

Không sửa tay file `internal/db/sqlc/*.go`. Sửa SQL rồi chạy `sqlc generate`.

## Tự demo

Chạy `scripts/demo.ps1`, sau đó dùng Postman thử thêm:

- Customer gọi API quản trị → 403.
- Thiếu token khi xem cart → 401.
- Tạo SKU/email trùng → 409.
- Nhập giá âm/số lượng 0 → 400.
- PUT product với version cũ → 409.
- Xem đơn của user khác → 404.
- Retry checkout cùng key → cùng order ID.
- Hủy đơn hai lần → stock chỉ tăng một lần.
- Admin chuyển processing rồi customer hủy → 409.

## Câu hỏi phỏng vấn

1. Vì sao service dùng context.Context thay vì gin.Context?
2. UNIQUE constraint khác kiểm tra tồn tại trước INSERT thế nào?
3. Vì sao cart mutex trong Go chưa đủ nếu chạy hai API instances?
4. Nếu request timeout sau khi commit, client retry sẽ làm gì?
5. Vì sao order items lưu snapshot giá?
6. FOR UPDATE và optimistic version phục vụ hai tình huống nào?
7. Vì sao phải dừng HTTP trước khi đóng pool?
8. Vì sao log request body trong bài học cần điều chỉnh khi dùng thực tế?
9. JWT payload có bí mật không? Vì sao không cần mã hóa AES cho MVP này?
10. Docker HEALTHCHECK khác readiness và CI khác CD ở điểm nào?

Mốc review: chạy demo, giải thích checkout/rollback/idempotency bằng lời của bạn, rồi mới thêm UI hoặc deploy domain. Không thêm công nghệ chỉ để tăng số dòng CV.
