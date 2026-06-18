# Product:

- id: UUID
- name: string
- description: string
- price: int64
- category: string
- deliveryDays: int
- rating: int32
- createdAt: time.Time
- updatedAt: time.Time

# Order:

- id: UUID
- userID: UUID
- status: NEW | PAID | DELIVERING | DELIVERED
- totalPrice: int64
- createdAt: time.Time
- deliveryDate: \*time.Time

# OrderItem:

- id: UUID
- orderID: UUID
- productID: UUID
- productName: string
- productPrice: int64
- quantity: int

# Cart:

- id: UUID
- userID: UUID
- createdAt: time.Time
- updatedAt: time.Time

# CartItem:

- id: UUID
- cartID: UUID
- productID: UUID
- quantity: int
