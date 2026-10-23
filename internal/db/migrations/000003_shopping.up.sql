CREATE TABLE carts (
 id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
 user_id INT NOT NULL UNIQUE REFERENCES users(user_id)
);
CREATE TABLE cart_items (
 cart_id BIGINT NOT NULL REFERENCES carts(id) ON DELETE CASCADE,
 product_id UUID NOT NULL REFERENCES products(id),
 quantity INT NOT NULL CHECK(quantity BETWEEN 1 AND 1000),
 PRIMARY KEY(cart_id,product_id)
);
CREATE TABLE orders (
 id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
 user_id INT NOT NULL REFERENCES users(user_id),
 status TEXT NOT NULL DEFAULT 'placed' CHECK(status IN ('placed','processing','shipped','completed','cancelled')),
 total_vnd BIGINT NOT NULL CHECK(total_vnd>=0),
 recipient_name VARCHAR(100) NOT NULL,
 phone VARCHAR(20) NOT NULL,
 address VARCHAR(500) NOT NULL,
 payment_method TEXT NOT NULL DEFAULT 'cod' CHECK(payment_method='cod'),
 idempotency_key UUID NOT NULL,
 request_hash TEXT NOT NULL,
 created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
 updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
 UNIQUE(user_id,idempotency_key)
);
CREATE TABLE order_items (
 order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
 product_id UUID NOT NULL REFERENCES products(id),
 sku VARCHAR(64) NOT NULL,
 product_name VARCHAR(200) NOT NULL,
 unit_price_vnd BIGINT NOT NULL CHECK(unit_price_vnd>0),
 quantity INT NOT NULL CHECK(quantity BETWEEN 1 AND 1000),
 PRIMARY KEY(order_id,product_id)
);
CREATE INDEX idx_orders_user ON orders(user_id,created_at DESC,id);
CREATE INDEX idx_orders_status ON orders(status,created_at DESC,id);
