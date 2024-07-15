package auth

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/justenwalker/mack/encoding/msgpack"
	"github.com/justenwalker/mack/macaroon"
	"github.com/justenwalker/mack/macaroon/thirdparty"
)

var _ thirdparty.ThirdParty = (*ThirdParty)(nil)

// ThirdParty implements thirdparty.ThirdParty for the Auth Service.
type ThirdParty struct {
	client      ClientWithResponsesInterface
	location    string
	credentials Credentials
	accessToken string
}

type Credentials struct {
	Username string
	Password string
}

func NewThirdParty(ctx context.Context, location string, credentials Credentials) (*ThirdParty, error) {
	client, err := NewClientWithResponses(location)
	if err != nil {
		return nil, err
	}
	tp := &ThirdParty{
		credentials: credentials,
		client:      client,
		location:    location,
	}
	if err = tp.login(ctx); err != nil {
		return nil, fmt.Errorf("failed to log into auth service: %w", err)
	}
	return tp, nil
}

// AccessToken returns the auth service access token used to communicate with the auth service.
func (t *ThirdParty) AccessToken() string {
	return t.accessToken
}

func (t *ThirdParty) MatchCaveat(c *macaroon.Caveat) bool {
	return c.Location() != t.location
}

func (t *ThirdParty) DischargeCaveat(ctx context.Context, c *macaroon.Caveat) (m macaroon.Macaroon, err error) {
	resp, err := t.client.PostDischargeWithResponse(ctx, PostDischargeJSONRequestBody{
		CaveatId: base64.StdEncoding.EncodeToString(c.ID()),
	}, func(ctx context.Context, req *http.Request) error {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", t.accessToken))
		return nil
	})
	if err != nil {
		return m, fmt.Errorf("discharge request failed: %w", err)
	}
	switch resp.StatusCode() {
	case http.StatusOK:
		var mbs []byte
		mbs, err = base64.StdEncoding.DecodeString(resp.JSON200.Macaroon)
		if err != nil {
			return m, fmt.Errorf("macaroon base64-decode failed: %w", err)
		}
		if err = msgpack.Encoding.DecodeMacaroon(mbs, &m); err != nil {
			return m, fmt.Errorf("macaroon unmarshal failed: %w", err)
		}
		return m, nil
	default:
		return m, fmt.Errorf("error %d: %s", resp.JSONDefault.Code, resp.JSONDefault.Error)
	}
}

func (t *ThirdParty) login(ctx context.Context) error {
	resp, err := t.client.PostLoginWithResponse(ctx, PostLoginJSONRequestBody{
		Username: t.credentials.Username,
		Password: t.credentials.Password,
	})
	if err != nil {
		return fmt.Errorf("login failed: %w", err)
	}
	switch resp.StatusCode() {
	case http.StatusOK:
		t.accessToken = resp.JSON200.AccessToken
		return nil
	default:
		return fmt.Errorf("login error: %d: %s", resp.JSONDefault.Code, resp.JSONDefault.Error)
	}
}
