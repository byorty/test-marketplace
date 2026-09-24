CREATE TABLE IF NOT EXISTS deliveries (
    id UUID NOT NULL PRIMARY KEY,
    order_id UUID NOT NULL UNIQUE,
    user_id UUID NOT NULL,
    status VARCHAR(30) NOT NULL
        CHECK (status IN ('PENDING', 'IN_TRANSIT', 'READY_FOR_PICKUP', 'DELIVERED', 'CANCELLED')),
    pickup_address TEXT NOT NULL
        CHECK (length(pickup_address) > 0),
    estimated_delivery_date TIMESTAMP NOT NULL,
    is_late BOOLEAN NOT NULL DEFAULT FALSE,
    qr_nonce UUID,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_deliveries_order_id ON deliveries(order_id);
CREATE INDEX idx_deliveries_user_id ON deliveries(user_id);
CREATE INDEX idx_deliveries_status ON deliveries(status);
CREATE INDEX idx_deliveries_is_late ON deliveries(is_late);

CREATE TABLE IF NOT EXISTS delivery_reschedules (
    id UUID NOT NULL PRIMARY KEY,
    delivery_id UUID NOT NULL
        REFERENCES deliveries(id) ON DELETE CASCADE,
    previous_date TIMESTAMP NOT NULL,
    new_date TIMESTAMP NOT NULL,
    reason TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_delivery_reschedules_delivery_id ON delivery_reschedules(delivery_id);