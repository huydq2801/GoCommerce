-- name: EnsureCart :one
INSERT INTO carts(user_id) VALUES($1) ON CONFLICT(user_id) DO UPDATE SET user_id=excluded.user_id RETURNING id;
-- name: ListCartItems :many
SELECT ci.product_id,ci.quantity,p.name,p.sku,p.price_vnd,p.stock_quantity,p.is_active FROM cart_items ci JOIN products p ON p.id=ci.product_id WHERE ci.cart_id=$1 ORDER BY ci.product_id;
-- name: CountCartItems :one
SELECT count(*) FROM cart_items WHERE cart_id=$1;
-- name: UpsertCartItem :exec
INSERT INTO cart_items(cart_id,product_id,quantity) VALUES($1,$2,$3) ON CONFLICT(cart_id,product_id) DO UPDATE SET quantity=excluded.quantity;
-- name: DeleteCartItem :exec
DELETE FROM cart_items WHERE cart_id=$1 AND product_id=$2;
-- name: ClearCart :exec
DELETE FROM cart_items WHERE cart_id=$1;
-- name: LockProduct :one
SELECT * FROM products WHERE id=$1 FOR UPDATE;
-- name: DecreaseStock :execrows
UPDATE products SET stock_quantity=stock_quantity-$2,version=version+1,updated_at=now() WHERE id=$1 AND stock_quantity >= $2 AND is_active;
-- name: RestoreStock :exec
UPDATE products SET stock_quantity=stock_quantity+$2,version=version+1,updated_at=now() WHERE id=$1;
-- name: CreateOrder :one
INSERT INTO orders(user_id,total_vnd,recipient_name,phone,address,idempotency_key,request_hash) VALUES($1,$2,$3,$4,$5,$6,$7) RETURNING *;
-- name: CreateOrderItem :exec
INSERT INTO order_items(order_id,product_id,sku,product_name,unit_price_vnd,quantity) VALUES($1,$2,$3,$4,$5,$6);
-- name: GetOrderByKey :one
SELECT * FROM orders WHERE user_id=$1 AND idempotency_key=$2;
-- name: GetOrder :one
SELECT * FROM orders WHERE id=$1;
-- name: LockOrder :one
SELECT * FROM orders WHERE id=$1 FOR UPDATE;
-- name: ListOrderItems :many
SELECT * FROM order_items WHERE order_id=$1 ORDER BY product_id;
-- name: ListOrders :many
SELECT * FROM orders WHERE (sqlc.arg(user_id)::int=0 OR user_id=sqlc.arg(user_id)) ORDER BY created_at DESC,id DESC LIMIT sqlc.arg(page_limit)::int OFFSET sqlc.arg(page_offset)::int;
-- name: CountOrders :one
SELECT count(*) FROM orders WHERE (sqlc.arg(user_id)::int=0 OR user_id=sqlc.arg(user_id));
-- name: SetOrderStatus :one
UPDATE orders SET status=$2,updated_at=now() WHERE id=$1 RETURNING *;
