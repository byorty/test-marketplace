package product

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
)

type HTTPClient struct {
	baseURL string
	client *http.Client
}

func NewClient(baseURL string) *HTTPClient {
	return &HTTPClient{
		baseURL: baseURL,
		client: &http.Client{},
	}
}

func (c *HTTPClient) GetProduct(ctx context.Context, id uuid.UUID) (*Product, error) {

	url := fmt.Sprintf("%s/products/%s", c.baseURL, id)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	switch resp.StatusCode {

	case http.StatusOK:

	case http.StatusNotFound:	
		return nil, ErrProductNotFound

	default:
		return nil, fmt.Errorf("product service returned %d", resp.StatusCode)
	}

	var product Product

	if err := json.NewDecoder(resp.Body).Decode(&product); err != nil {
		return nil, err
	}

	return &product, nil
}