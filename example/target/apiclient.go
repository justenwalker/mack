package target

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"

	"example/headers"

	"github.com/justenwalker/mack/encoding/msgpack"
	"github.com/justenwalker/mack/macaroon"
)

type APIClient struct {
	accessToken string
	client      ClientWithResponsesInterface
}

func NewAPIClient(location string, accessToken string) (*APIClient, error) {
	client, err := NewClientWithResponses(location)
	if err != nil {
		return nil, fmt.Errorf("NewClientWithResponses: %w", err)
	}
	return &APIClient{
		accessToken: accessToken,
		client:      client,
	}, nil
}

func (c *APIClient) GetMacaroon(ctx context.Context, org string, app string) (m macaroon.Macaroon, err error) {
	resp, err := c.client.GetMacaroonRequestWithResponse(ctx, &GetMacaroonRequestParams{
		Org: ptr(org),
		App: ptr(app),
	}, func(ctx context.Context, req *http.Request) error {
		req.Header.Set("Authorization", "Bearer "+c.accessToken)
		return nil
	})
	if err != nil {
		return m, fmt.Errorf("login failed: %w", err)
	}
	switch resp.StatusCode() {
	case http.StatusOK:
		var mbs []byte
		mbs, err = base64.StdEncoding.DecodeString(resp.JSON200.Macaroon)
		if err != nil {
			return m, fmt.Errorf("base64 decode failed: %w", err)
		}
		err = msgpack.Encoding.DecodeMacaroon(mbs, &m)
		if err != nil {
			return m, fmt.Errorf("decode macaroon failed: %w", err)
		}
		return m, nil
	default:
		return m, fmt.Errorf("get-macaroon failed: %d: %s", resp.JSONDefault.Code, resp.JSONDefault.Error)
	}
}

type Operation struct {
	Org       string
	App       string
	Operation string
	Args      []string
}

func (c *APIClient) DoOperation(ctx context.Context, stack macaroon.Stack, op Operation) (map[string]interface{}, error) {
	var args *[]string
	if len(op.Args) == 0 {
		args = &op.Args
	}
	resp, err := c.client.PostOrgAppDoWithResponse(ctx, op.Org, op.App, PostOrgAppDoJSONRequestBody{
		Arguments: args,
		Operation: op.Operation,
	}, func(ctx context.Context, req *http.Request) error {
		headers.SetMacaroonStackAuthorization(req.Header, stack)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("do operation failed: %w", err)
	}
	switch resp.StatusCode() {
	case http.StatusOK:
		return *resp.JSON200, nil
	default:
		return nil, fmt.Errorf("do operation failed: %d: %s", resp.JSONDefault.Code, resp.JSONDefault.Error)
	}
}

func ptr[T any](v T) *T {
	return &v
}
