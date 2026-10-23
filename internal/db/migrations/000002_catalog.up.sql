CREATE TABLE categories (
 id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
 name VARCHAR(100) NOT NULL CHECK (length(trim(name)) > 0),
 slug VARCHAR(100) NOT NULL UNIQUE CHECK (slug ~ '^[a-z0-9]+(-[a-z0-9]+)*$'),
 created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
 updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE products (
 id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
 category_id UUID NOT NULL REFERENCES categories(id) ON DELETE RESTRICT,
 sku VARCHAR(64) NOT NULL UNIQUE,
 name VARCHAR(200) NOT NULL CHECK (length(trim(name)) > 0),
 slug VARCHAR(200) NOT NULL UNIQUE CHECK (slug ~ '^[a-z0-9]+(-[a-z0-9]+)*$'),
 brand VARCHAR(100) NOT NULL CHECK (length(trim(brand)) > 0),
 model VARCHAR(100) NOT NULL,
 description TEXT NOT NULL DEFAULT '',
 price_vnd BIGINT NOT NULL CHECK (price_vnd > 0 AND price_vnd <= 1000000000000),
 stock_quantity INT NOT NULL DEFAULT 0 CHECK (stock_quantity >= 0),
 specifications JSONB NOT NULL DEFAULT '{}'::jsonb CHECK (jsonb_typeof(specifications) = 'object'),
 is_active BOOLEAN NOT NULL DEFAULT true,
 version BIGINT NOT NULL DEFAULT 1,
 created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
 updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_products_category ON products(category_id);
CREATE INDEX idx_products_public_price ON products(price_vnd,id) WHERE is_active;
CREATE INDEX idx_products_brand ON products(lower(brand)) WHERE is_active;
