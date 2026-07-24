package product

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

var (
	ErrProductNotFound = errors.New("product not found")
)

type Client interface {
	GetByID(ctx context.Context, id uuid.UUID) (*Product, error)
}

type client struct {
	httpClient *http.Client
	baseURL string
}

type ProductService struct {
	URL string `env:"PRODUCT_SERVICE_URL"`
}

func New(baseURL string) Client {
	return &client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (c *client) GetByID(ctx context.Context, id uuid.UUID) (*Product, error) {
	req, err := http.NewRequestWithContext(
		ctx, http.MethodGet, fmt.Sprintf("%s/products/%s", c.baseURL, id), nil,
	)

	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	switch resp.StatusCode{
	case http.StatusOK:
		var product Product

		if err := json.NewDecoder(resp.Body).Decode(&product); err != nil {
			return nil, err
		}

		return &product, nil

	case http.StatusNotFound:
		return nil, ErrProductNotFound

	default: 
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
}