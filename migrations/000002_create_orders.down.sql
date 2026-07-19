DROP TABLE IF EXISTS orders;

DROP TABLE IF EXISTS order_items;

DROP TABLE IF EXISTS cart_items;

DROP INDEX idx_orders_user_id;

DROP INDEX idx_orders_status;

DROP INDEX idx_order_items_order_id;

DROP INDEX idx_cart_items_user_id;
