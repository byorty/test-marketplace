DROP TABLE IF EXISTS order_items;

DROP TABLE IF EXISTS cart_items;

DROP TABLE IF EXISTS orders;

DROP INDEX IF EXISTS idx_orders_user_id;

DROP INDEX IF EXISTS idx_orders_status;

DROP INDEX IF EXISTS idx_order_items_order_id;

DROP INDEX IF EXISTS idx_cart_items_user_id;