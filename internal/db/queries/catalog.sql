-- name: CreateCategory :one
INSERT INTO categories(name,slug) VALUES($1,$2) RETURNING *;
-- name: ListCategories :many
SELECT * FROM categories ORDER BY name,id;
-- name: UpdateCategory :one
UPDATE categories SET name=$2,slug=$3,updated_at=now() WHERE id=$1 RETURNING *;
-- name: DeleteCategory :execrows
DELETE FROM categories WHERE id=$1;
-- name: CreateProduct :one
INSERT INTO products(category_id,sku,name,slug,brand,model,description,price_vnd,stock_quantity,specifications)
VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) RETURNING *;
-- name: GetProduct :one
SELECT * FROM products WHERE id=$1 AND is_active;
-- name: GetAdminProduct :one
SELECT * FROM products WHERE id=$1;
-- name: UpdateProduct :one
UPDATE products SET category_id=sqlc.arg(category_id),sku=sqlc.arg(sku),name=sqlc.arg(name),slug=sqlc.arg(slug),brand=sqlc.arg(brand),model=sqlc.arg(model),description=sqlc.arg(description),price_vnd=sqlc.arg(price_vnd),stock_quantity=sqlc.arg(stock_quantity),specifications=sqlc.arg(specifications),is_active=sqlc.arg(is_active),version=version+1,updated_at=now()
WHERE id=sqlc.arg(id) AND version=sqlc.arg(expected_version) RETURNING *;
-- name: DeactivateProduct :execrows
UPDATE products SET is_active=false,version=version+1,updated_at=now() WHERE id=$1 AND is_active;
-- name: ListProducts :many
SELECT * FROM products WHERE is_active
AND (sqlc.arg(search)::text='' OR name ILIKE '%'||sqlc.arg(search)||'%' OR model ILIKE '%'||sqlc.arg(search)||'%')
AND (sqlc.arg(category_id)::uuid='00000000-0000-0000-0000-000000000000'::uuid OR category_id=sqlc.arg(category_id))
AND (sqlc.arg(brand)::text='' OR lower(brand)=lower(sqlc.arg(brand)))
AND (sqlc.arg(min_price)::bigint<0 OR price_vnd>=sqlc.arg(min_price))
AND (sqlc.arg(max_price)::bigint<0 OR price_vnd<=sqlc.arg(max_price))
ORDER BY CASE WHEN sqlc.arg(sort_by)::text='price_asc' THEN price_vnd END ASC,
 CASE WHEN sqlc.arg(sort_by)::text='price_desc' THEN price_vnd END DESC,
 created_at DESC,id DESC
LIMIT sqlc.arg(page_limit)::int OFFSET sqlc.arg(page_offset)::int;
-- name: CountProducts :one
SELECT count(*) FROM products WHERE is_active
AND (sqlc.arg(search)::text='' OR name ILIKE '%'||sqlc.arg(search)||'%' OR model ILIKE '%'||sqlc.arg(search)||'%')
AND (sqlc.arg(category_id)::uuid='00000000-0000-0000-0000-000000000000'::uuid OR category_id=sqlc.arg(category_id))
AND (sqlc.arg(brand)::text='' OR lower(brand)=lower(sqlc.arg(brand)))
AND (sqlc.arg(min_price)::bigint<0 OR price_vnd>=sqlc.arg(min_price))
AND (sqlc.arg(max_price)::bigint<0 OR price_vnd<=sqlc.arg(max_price));
