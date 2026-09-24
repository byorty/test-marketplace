package order

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
)

type OrderResponse struct {
	ID     uuid.UUID `json:"id"`
	UserID uuid.UUID `json:"user_id"`
	Status string    `json:"status"`
}

type OrderClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewOrderClient(baseURL string) *OrderClient {
	return &OrderClient{
		baseURL:    baseURL,
		httpClient: &http.Client{},
	}
}

func (c *OrderClient) GetOrderByID(ctx context.Context, orderID uuid.UUID) (*OrderResponse, error) {
	url := fmt.Sprintf("%s/api/v1/orders/%s", c.baseURL, orderID.String())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("get order request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrOrderNotFound
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("order service returned status %d", resp.StatusCode)
	}

	var order OrderResponse
	if err := json.NewDecoder(resp.Body).Decode(&order); err != nil {
		return nil, fmt.Errorf("decode order response: %w", err)
	}

	return &order, nil
}

func (c *OrderClient) UpdateOrderStatus(ctx context.Context, orderID uuid.UUID, status string) error {
	return ErrNotImplemented
}
