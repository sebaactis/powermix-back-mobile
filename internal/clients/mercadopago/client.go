package mercadopago

import "net/http"

type Client struct {
	Token  string
	Client *http.Client
}

func NewClient(token string) *Client {
	return &Client{
		Token:  token,
		Client: &http.Client{},
	}
}
