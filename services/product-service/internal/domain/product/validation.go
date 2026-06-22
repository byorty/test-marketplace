package product

func (p *Product) Validate() error {
	if p.Name == "" {
		return ErrInvalidName
	}

	if p.Price < 0 {
		return ErrInvalidPrice
	}

	if p.DeliveryDays < 1 {
		return ErrInvalidDeliveryDays
	}

	if p.Rating < 0 || p.Rating > 5 {
		return ErrInvalidRating
	}

	return nil
}
