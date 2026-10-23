# Development guide

## Luồng khởi động

`cmd/api/main.go` đọc cấu hình, mở PostgreSQL pool và khởi tạo application. `internal/app` nối repository, service, handler và router bằng constructor. Khi nhận tín hiệu dừng, HTTP server drain các request trước khi pool được đóng.

## Tổ chức nghiệp vụ

- **Handler:** đọc request, validate input, gọi service và ánh xạ response HTTP.
- **Service:** áp dụng quy tắc nghiệp vụ, quyền sở hữu và ranh giới transaction.
- **Repository:** cung cấp truy cập dữ liệu và transaction store.
- **sqlc:** sinh mã Go từ SQL trong `internal/db/queries`.

Service nhận `context.Context`; các truy vấn dùng context của request để hỗ trợ cancellation và deadline.

## Thay đổi SQL và schema

1. Tạo migration mới khi thay đổi schema. Không sửa migration đã áp dụng trên môi trường có dữ liệu.
2. Cập nhật SQL trong `internal/db/queries`.
3. Chạy `sqlc generate`; không sửa tay `internal/db/sqlc/*.go`.
4. Cập nhật repository/service/handler và OpenAPI.
5. Bổ sung test cho hành vi hoặc ràng buộc mới.

## Checkout và inventory

Mọi sửa giỏ và checkout khóa cùng cart row. Checkout khóa các product theo UUID tăng dần; kiểm tra active/stock, lấy giá hiện tại và lưu snapshot trong cùng transaction với việc trừ kho và xóa giỏ.

Giữ thứ tự khóa nhất quán khi bổ sung thao tác inventory. Không thay row locks bằng mutex trong process vì ứng dụng có thể chạy nhiều instance. Các thao tác thay đổi product tăng version để phát hiện form quản trị đã lỗi thời.

Idempotency key định danh một checkout. Retry cùng key phải trả đơn cũ; thay đổi input với key đã dùng phải bị từ chối.

## Quy ước API

- Public đọc catalog; customer JWT truy cập giỏ và đơn của chính tài khoản đó.
- Admin JWT hoặc bootstrap credential mới được gọi API quản trị.
- Không nhận role/status trong payload đăng ký công khai.
- Dùng allowlist cho sort, giới hạn pagination và kích thước request.
- Không đưa SQL errors, password, token hoặc nguyên request body vào response/log.
- Giữ OpenAPI trong `internal/apidocs/openapi.json` đồng bộ với route và DTO.

## Trước khi commit

```powershell
gofmt -w cmd internal pkg tests
go vet ./...
go test ./...
```

Chạy integration tests khi thay đổi query, migration, transaction hoặc API contract. Xem [Testing](verification.md) để chuẩn bị PostgreSQL test riêng.
