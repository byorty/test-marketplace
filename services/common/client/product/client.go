package product

import (
	"context"
	"encoding/json"
	"fmt"

	generated "github.com/byorty/test-marketplace/services/common/client/product/generated"
	"github.com/google/uuid"
)

type ProductClient struct {
	client *generated.Client
}

func NewProductClient() (*ProductClient, error) {
	c, err := generated.NewClient(
		"http://product-service:8080",
	)
	if err != nil {
		return nil, fmt.Errorf("create product client: %w", err)
	}

	return &ProductClient{
		client: c,
	}, nil
}

func (c *ProductClient) GetProduct(
	ctx context.Context,
	id uuid.UUID,
) (*generated.ProductResponse, error) {

	resp, err := c.client.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get product request: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf(
			"product service returned status %d",
			resp.StatusCode,
		)
	}

	var product generated.ProductResponse

	if err := json.NewDecoder(resp.Body).Decode(&product); err != nil {
		return nil, fmt.Errorf(
			"decode product response: %w",
			err,
		)
	}

	return &product, nil
}