CREATE TABLE IF NOT EXISTS products (
    id UUID NOT NULL PRIMARY KEY,

    name TEXT NOT NULL 
        CHECK(length(name) > 0),

    description TEXT NOT NULL,

    category TEXT NOT NULL,

    price BIGINT NOT NULL
        CHECK (price >= 0),
        
    delivery_days INTEGER NOT NULL 
        CHECK (delivery_days >= 1),

    rating DECIMAL(3,2) NOT NULL DEFAULT 0
        CHECK (rating >= 0 AND rating <= 5),

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),

    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_products_name
ON products (name);

CREATE INDEX idx_products_category
ON products(category);

CREATE INDEX idx_products_price
ON products(price);

CREATE INDEX idx_products_rating
ON products(rating);

CREATE INDEX idx_products_delivery_days
ON products(delivery_days);
