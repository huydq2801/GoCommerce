# Dùng domain doanquanghuy.id.vn

Domain không thay VPS. Cấu hình này được chuẩn bị cho một Linux VPS do bạn quản lý; chưa tạo server, đổi DNS hay cấp certificate.

## Bố trí đề xuất

- `api.doanquanghuy.id.vn`: GoCommerce backend và `/openapi.json`.
- `doanquanghuy.id.vn`: portfolio/trang giới thiệu hoặc storefront ở giai đoạn sau.
- PostgreSQL chỉ bind loopback host; không đưa cổng 5432 ra Internet.
- Nginx chạy trên host, proxy tới `127.0.0.1:8080`; Compose API cũng chỉ publish loopback.

## Các bước triển khai khi đã có VPS

1. Chuẩn bị Linux, Docker Engine/Compose, Nginx và công cụ ACME như Certbot. Tạo user deploy không dùng root cho thao tác thường ngày.
2. Trong DNS zone hiện tại, thêm A record `api` trỏ IP public của VPS. Không đổi record root/mail đang dùng. Chỉ thêm AAAA nếu VPS phục vụ IPv6 đúng.
3. Đưa source lên VPS bằng repository/private transfer đã được bạn duyệt; tạo `.env` trên server và hạn chế quyền file bằng `chmod 600 .env`.
4. Dùng secret ngẫu nhiên độc lập cho DB_PASSWORD, API_KEY và JWT_SECRET; không dùng dữ liệu ví dụ/test. Với Compose migrate URL hiện tại, dùng password dạng hex để tránh ký tự URL đặc biệt.
5. Backup database hiện có, chạy `docker compose up --build --wait`, kiểm tra `curl http://127.0.0.1:8080/health/ready`.
6. Cài template `deploy/nginx/api.conf`, chạy `sudo nginx -t` rồi reload Nginx.
7. Khi DNS đã trỏ đúng, cấp certificate cho `api.doanquanghuy.id.vn`, ví dụ `sudo certbot --nginx -d api.doanquanghuy.id.vn`. Mở inbound 80/443 và giữ cổng SSH đang dùng. Không bật firewall mới trước khi giữ đường SSH của bạn.
8. Kiểm tra HTTPS, redirect HTTP → HTTPS và lịch tự gia hạn; thử `sudo certbot renew --dry-run`.
9. Chạy demo qua HTTPS và chỉ lúc đó mới thêm URL demo vào CV.

HTTP-01 cần port 80 truy cập được từ ngoài. Nếu có CDN/proxy khác, cần điều chỉnh trust chain và ACME theo mô hình thực tế; template hiện dành cho một Nginx trực tiếp phía trước API.

## IP và rate limit sau Nginx

Nginx template ghi đè X-Forwarded-For bằng IP kết nối trực tiếp, tránh tin header khách tự gửi. Go mặc định không tin proxy nào. Khi deploy, xác định source IP mà API nhìn thấy từ Docker gateway/Nginx rồi đặt `TRUSTED_PROXIES` bằng đúng IP/CIDR nhỏ đó; có thể xem gateway bằng `docker inspect` trên container API. Không dùng `0.0.0.0/0` hoặc `::/0`.

Nếu chưa cấu hình, nhiều khách qua Nginx có thể bị gom cùng một IP trong rate limiter. Đừng sửa bằng cách tắt rate limit hoặc tin mọi proxy.

## Backup và rollback

Tạo thư mục backup ngoài Git, hạn chế quyền file. Ví dụ trên Linux, dùng biến môi trường sẵn có trong container để dump, không đưa password vào history:

```sh
mkdir -p backups
chmod 700 backups
docker compose exec -T db sh -c 'pg_dump -U "$POSTGRES_USER" -d "$POSTGRES_DB" -Fc' > backups/gocommerce.dump
chmod 600 backups/gocommerce.dump
```

Để xác nhận backup dùng được, restore file vào một database **riêng**, không thử đè lên database đang chạy:

```sh
docker compose exec -T db sh -c 'createdb -U "$POSTGRES_USER" gocommerce_restore_test'
docker compose exec -T db sh -c 'pg_restore -U "$POSTGRES_USER" -d gocommerce_restore_test --exit-on-error' < backups/gocommerce.dump
```

Kiểm tra bảng/số lượng bản ghi trong database restore trước khi coi backup là đạt. Các lệnh trên là hướng dẫn, chưa được chạy với dữ liệu của bạn.

Giữ lại image trước khi deploy. Để rollback API, dùng image đã build trước đó tương thích schema; không mặc định chạy down migration vì có thể xóa bảng và dữ liệu. Production ưu tiên migration bổ sung để sửa schema. Đừng dùng migrate force để che lỗi chưa hiểu.

## CI/CD

CI workflow đã có cấu hình format/vet/test/build/Docker. Chưa có CD, remote SSH job hoặc auto deployment. Chỉ thêm CD sau lần deploy thủ công thành công và khi bạn duyệt quyền truy cập VPS/secrets.

Nguồn tham khảo: [Nginx proxy module](https://nginx.org/en/docs/http/ngx_http_proxy_module.html), [Let's Encrypt HTTP-01 và port 80](https://letsencrypt.org/docs/allow-port-80/), [Compose startup order](https://docs.docker.com/compose/how-tos/startup-order/).
