package coffeeji

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type Client struct {
	baseURL    string
	key        string
	secret     string
	httpClient *http.Client
}

func NewClient(key, secret string) *Client {
	return &Client{
		baseURL: "https://gsvden.coffeeji.com",
		key:     key,
		secret:  secret,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Client) GetOrderList(ctx context.Context, orderNo string) (*OrderListResponse, error) {
	u, err := url.Parse(c.baseURL + "/coffee/newThird/order/getOrderList")
	if err != nil {
		return nil, fmt.Errorf("parse base url: %w", err)
	}

	q := u.Query()
	q.Set("orderNo", orderNo)
	u.RawQuery = q.Encode()

	timestamp, keyMd5 := buildAuthHeaders(c.key, c.secret)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("crear request: %w", err)
	}

	req.Header.Set("key", c.key)
	req.Header.Set("key-md5", keyMd5)
	req.Header.Set("timestamp", timestamp)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("llamando endpoint getOrderList: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("leyendo body: %w", err)
	}

	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("status no exitoso: %s, body: %s", resp.Status, string(body))
	}

	var parsed OrderListResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, fmt.Errorf("parseando json: %w", err)
	}

	return &parsed, nil
}

func (c *Client) GetGoodsNameByOrderNo(ctx context.Context, orderNo string) (string, error) {
	resp, err := c.GetOrderList(ctx, orderNo)
	if err != nil {
		return "", err
	}

	if !resp.Success || resp.Code != 200 {
		return "", fmt.Errorf("respuesta no exitosa: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	if len(resp.Data.Records) == 0 {
		return "", fmt.Errorf("no se encontraron registros para orderNo %s", orderNo)
	}

	return resp.Data.Records[0].GoodsName, nil
}
