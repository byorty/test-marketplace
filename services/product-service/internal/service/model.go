package service

type CreateProduct struct {
	Name string
	Description string
	Category string
	Price int64
	DeliveryDays int
}

type UpdateProduct struct {
	Name *string
	Description *string
	Category *string
	Price *int64
	DeliveryDays *int
}

