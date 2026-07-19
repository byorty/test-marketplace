CREATE TABLE IF NOT EXISTS orders (
    id UUID NOT NULL PRIMARY KEY,

    user_id UUID NOT NULL,

    status TEXT NOT NULL
        CHECK (
            status IN (
                'CREATED',
                'PAID',
                'DELIVERING',
                'DELIVERED'
            )
        ),

    total_price BIGINT NOT NULL
        CHECK (total_price >= 0),

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),

    delivery_date TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS order_items (
    id UUID NOT NULL PRIMARY KEY,

    order_id UUID NOT NULL
        REFERENCES orders(id)
        ON DELETE CASCADE,

    product_id UUID NOT NULL,

    product_name TEXT NOT NULL
        CHECK (length(product_name) > 0),

    product_price BIGINT NOT NULL
        CHECK (product_price >= 0),

    quantity INTEGER NOT NULL
        CHECK (quantity > 0)
);

CREATE TABLE IF NOT EXISTS cart_items (
    id UUID NOT NULL PRIMARY KEY,

    user_id UUID NOT NULL,

    product_id UUID NOT NULL,

    quantity INTEGER NOT NULL
        CHECK (quantity > 0),

    UNIQUE(user_id, product_id)
);

CREATE INDEX idx_orders_user_id 
ON orders(user_id);

CREATE INDEX idx_orders_status
ON orders(status);

CREATE INDEX idx_order_items_order_id
ON order_items(order_id);

CREATE INDEX idx_cart_items_user_id
ON cart_items(user_id);

