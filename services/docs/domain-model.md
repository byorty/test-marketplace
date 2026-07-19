# Product:

- id: UUID
- name: string
- description: string
- price: int64
- category: string
- delivery_days: int
- rating: float64
- created_at: time.Time
- updated_at: time.Time

# Order:

- id: UUID
- user_ID: UUID
- status: NEW | PAID | DELIVERING | DELIVERED
- total_price: int64
- createdAt: time.Time
- delivery_date: time.Time

# OrderItems:

- id: UUID
- order_ID: UUID
- product_ID: UUID
- product_name: string
- product_price: int64
- quantity: int

# CartItems:

- id: UUID
- user_ID: UUID
- product_ID: UUID
- quantity: int
