package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)


func GetGoogleUserInfo(ctx context.Context, accessToken string) (*OAuthUserInfo, error) {
	req, _ := http.NewRequestWithContext(ctx, "GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil)

	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token de Google invalido")
	}

	var data struct {
		ID      string `json:"id"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	return &OAuthUserInfo{
		Provider:   "google",
		ProviderID: data.ID,
		Email:      data.Email,
		Name:       data.Name,
		PictureURL: data.Picture,
	}, nil
}
