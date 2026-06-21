package product

type SortBy string 

const (
	SortByPrice SortBy = "price"
	SortByRating SortBy = "rating"
)

type SortOrder string 

const (
	Asc SortOrder = "asc"
	Desc SortOrder = "desc"
)

type ListFilter struct {
	Name string
	Category string
	MinPrice *int64
	MaxPrice *int64
	MinRating *float64
	DeliveryDays *int

	SortBy SortBy
	Order SortOrder

	Page int
	PageSize int
}

type ProductList struct {
	Items []Product
	Total int64
	Page int
	PageSize int 
}