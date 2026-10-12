package line

import (
	"bakery/internal/domains/users"
	"bakery/internal/shared"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
)

type Identity struct {
	Client    *http.Client
	ChannelID string
}

func (i Identity) Verify(ctx context.Context, token string) (users.Profile, error) {
	if i.ChannelID == "" || token == "" {
		return users.Profile{}, shared.ErrForbidden
	}
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.line.me/oauth2/v2.1/verify?access_token="+url.QueryEscape(token), nil)
	if err != nil {
		return users.Profile{}, err
	}
	res, err := i.Client.Do(req)
	if err != nil {
		return users.Profile{}, shared.ErrForbidden
	}
	var v struct {
		ChannelID string `json:"client_id"`
		Expiry    int    `json:"expires_in"`
	}
	err = json.NewDecoder(io.LimitReader(res.Body, 64<<10)).Decode(&v)
	_ = res.Body.Close()
	if err != nil || res.StatusCode != 200 || v.ChannelID != i.ChannelID || v.Expiry <= 0 {
		return users.Profile{}, shared.ErrForbidden
	}
	req, err = http.NewRequestWithContext(ctx, "GET", "https://api.line.me/v2/profile", nil)
	if err != nil {
		return users.Profile{}, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	res, err = i.Client.Do(req)
	if err != nil {
		return users.Profile{}, shared.ErrForbidden
	}
	defer func() { _ = res.Body.Close() }()
	var p struct {
		ID   string `json:"userId"`
		Name string `json:"displayName"`
	}
	err = json.NewDecoder(io.LimitReader(res.Body, 64<<10)).Decode(&p)
	if err != nil || res.StatusCode != 200 || p.ID == "" {
		return users.Profile{}, shared.ErrForbidden
	}
	return users.Profile{ID: p.ID, DisplayName: p.Name}, nil
}
