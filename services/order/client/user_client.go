package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
)

type UserSummary struct {
	ID    uuid.UUID `json:"id"`
	Email string    `json:"email"`
	Name  string    `json:"name"`
}

type UserClient interface {
	SearchUsers(ctx context.Context, query string, limit int) ([]UserSummary, error)
	GetUsersByIDs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]UserSummary, error)
}

type userClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewUserClient(baseURL string) UserClient {
	return &userClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type userListResponse struct {
	Success bool `json:"success"`
	Data    struct {
		Items []UserSummary `json:"items"`
		Total int           `json:"total"`
	} `json:"data"`
	Message string `json:"message"`
}

func (c *userClient) SearchUsers(ctx context.Context, query string, limit int) ([]UserSummary, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return []UserSummary{}, nil
	}

	params := url.Values{}
	params.Set("query", query)
	params.Set("limit", fmt.Sprintf("%d", limit))

	result, err := c.getUsers(ctx, "/internal/users/search?"+params.Encode())
	if err != nil {
		return nil, err
	}
	return result.Data.Items, nil
}

func (c *userClient) GetUsersByIDs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]UserSummary, error) {
	usersByID := make(map[uuid.UUID]UserSummary)
	if len(ids) == 0 {
		return usersByID, nil
	}

	idStrings := make([]string, 0, len(ids))
	for _, id := range ids {
		idStrings = append(idStrings, id.String())
	}

	params := url.Values{}
	params.Set("ids", strings.Join(idStrings, ","))

	result, err := c.getUsers(ctx, "/internal/users?"+params.Encode())
	if err != nil {
		return nil, err
	}
	for _, user := range result.Data.Items {
		usersByID[user.ID] = user
	}
	return usersByID, nil
}

func (c *userClient) getUsers(ctx context.Context, path string) (*userListResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create user-service request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call user-service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("user-service returned status %d", resp.StatusCode)
	}

	var result userListResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode user-service response: %w", err)
	}
	return &result, nil
}
