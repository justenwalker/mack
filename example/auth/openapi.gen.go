//go:build go1.22

// Package auth provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/oapi-codegen/oapi-codegen/v2 version v2.4.1 DO NOT EDIT.
package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	BearerHTTPAuthenticationScopes = "bearerHTTPAuthentication.Scopes"
)

// DischargeMacaroonRequestBody defines model for DischargeMacaroonRequestBody.
type DischargeMacaroonRequestBody struct {
	// CaveatId The caveat ID to be discharged
	CaveatId string `json:"caveat_id"`
}

// DischargeMacaroonResponseBody defines model for DischargeMacaroonResponseBody.
type DischargeMacaroonResponseBody struct {
	// ExpiresIn number of seconds for which this macaroon is valid after this response was generated
	ExpiresIn int64 `json:"expires_in"`

	// Macaroon A macaroon discharging the provided caveat ID
	Macaroon string `json:"macaroon"`
}

// ErrorResponseBody defines model for ErrorResponseBody.
type ErrorResponseBody struct {
	// Code error code
	Code int `json:"code"`

	// Error error message
	Error string `json:"error"`
}

// IdentitiesResponseBody defines model for IdentitiesResponseBody.
type IdentitiesResponseBody = []struct {
	// KeyId Key Identifier
	KeyId string `json:"key_id"`

	// KeyType Key Type
	KeyType string `json:"key_type"`

	// PublicKey Base-64 Encoded Public Key Data
	PublicKey string `json:"public_key"`
}

// LoginRequestBody defines model for LoginRequestBody.
type LoginRequestBody struct {
	Password string `json:"password"`
	Username string `json:"username"`
}

// LoginResponseBody defines model for LoginResponseBody.
type LoginResponseBody struct {
	// AccessToken The access token authorizing the user
	AccessToken string `json:"access_token"`

	// ExpiresIn number of seconds for which this macaroon is valid after this response was generated
	ExpiresIn int64 `json:"expires_in"`
}

// ValidateTokenRequestBody defines model for ValidateTokenRequestBody.
type ValidateTokenRequestBody struct {
	// AccessToken The access token authorizing the user
	AccessToken string `json:"access_token"`
}

// ValidateTokenResponseBody defines model for ValidateTokenResponseBody.
type ValidateTokenResponseBody struct {
	Expires  time.Time `json:"expires"`
	Username string    `json:"username"`
}

// DischargeMacaroonResponse defines model for DischargeMacaroonResponse.
type DischargeMacaroonResponse = DischargeMacaroonResponseBody

// ErrorResponse defines model for ErrorResponse.
type ErrorResponse = ErrorResponseBody

// IdentitiesResponse defines model for IdentitiesResponse.
type IdentitiesResponse = IdentitiesResponseBody

// LoginResponse defines model for LoginResponse.
type LoginResponse = LoginResponseBody

// ValidateTokenResponse defines model for ValidateTokenResponse.
type ValidateTokenResponse = ValidateTokenResponseBody

// DischargeMacaroonRequest defines model for DischargeMacaroonRequest.
type DischargeMacaroonRequest = DischargeMacaroonRequestBody

// LoginRequest defines model for LoginRequest.
type LoginRequest = LoginRequestBody

// ValidateTokenRequest defines model for ValidateTokenRequest.
type ValidateTokenRequest = ValidateTokenRequestBody

// PostDischargeJSONRequestBody defines body for PostDischarge for application/json ContentType.
type PostDischargeJSONRequestBody = DischargeMacaroonRequestBody

// PostLoginJSONRequestBody defines body for PostLogin for application/json ContentType.
type PostLoginJSONRequestBody = LoginRequestBody

// PostValidateTokenJSONRequestBody defines body for PostValidateToken for application/json ContentType.
type PostValidateTokenJSONRequestBody = ValidateTokenRequestBody

// RequestEditorFn  is the function signature for the RequestEditor callback function
type RequestEditorFn func(ctx context.Context, req *http.Request) error

// Doer performs HTTP requests.
//
// The standard http.Client implements this interface.
type HttpRequestDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client which conforms to the OpenAPI3 specification for this service.
type Client struct {
	// The endpoint of the server conforming to this interface, with scheme,
	// https://api.deepmap.com for example. This can contain a path relative
	// to the server, such as https://api.deepmap.com/dev-test, and all the
	// paths in the swagger spec will be appended to the server.
	Server string

	// Doer for performing requests, typically a *http.Client with any
	// customized settings, such as certificate chains.
	Client HttpRequestDoer

	// A list of callbacks for modifying requests which are generated before sending over
	// the network.
	RequestEditors []RequestEditorFn
}

// ClientOption allows setting custom parameters during construction
type ClientOption func(*Client) error

// Creates a new Client, with reasonable defaults
func NewClient(server string, opts ...ClientOption) (*Client, error) {
	// create a client with sane default values
	client := Client{
		Server: server,
	}
	// mutate client and add all optional params
	for _, o := range opts {
		if err := o(&client); err != nil {
			return nil, err
		}
	}
	// ensure the server URL always has a trailing slash
	if !strings.HasSuffix(client.Server, "/") {
		client.Server += "/"
	}
	// create httpClient, if not already present
	if client.Client == nil {
		client.Client = &http.Client{}
	}
	return &client, nil
}

// WithHTTPClient allows overriding the default Doer, which is
// automatically created using http.Client. This is useful for tests.
func WithHTTPClient(doer HttpRequestDoer) ClientOption {
	return func(c *Client) error {
		c.Client = doer
		return nil
	}
}

// WithRequestEditorFn allows setting up a callback function, which will be
// called right before sending the request. This can be used to mutate the request.
func WithRequestEditorFn(fn RequestEditorFn) ClientOption {
	return func(c *Client) error {
		c.RequestEditors = append(c.RequestEditors, fn)
		return nil
	}
}

// The interface specification for the client above.
type ClientInterface interface {
	// PostDischargeWithBody request with any body
	PostDischargeWithBody(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	PostDischarge(ctx context.Context, body PostDischargeJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)

	// GetIdentities request
	GetIdentities(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error)

	// PostLoginWithBody request with any body
	PostLoginWithBody(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	PostLogin(ctx context.Context, body PostLoginJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)

	// PostValidateTokenWithBody request with any body
	PostValidateTokenWithBody(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	PostValidateToken(ctx context.Context, body PostValidateTokenJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)
}

func (c *Client) PostDischargeWithBody(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewPostDischargeRequestWithBody(c.Server, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) PostDischarge(ctx context.Context, body PostDischargeJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewPostDischargeRequest(c.Server, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) GetIdentities(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewGetIdentitiesRequest(c.Server)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) PostLoginWithBody(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewPostLoginRequestWithBody(c.Server, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) PostLogin(ctx context.Context, body PostLoginJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewPostLoginRequest(c.Server, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) PostValidateTokenWithBody(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewPostValidateTokenRequestWithBody(c.Server, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) PostValidateToken(ctx context.Context, body PostValidateTokenJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewPostValidateTokenRequest(c.Server, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

// NewPostDischargeRequest calls the generic PostDischarge builder with application/json body
func NewPostDischargeRequest(server string, body PostDischargeJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewPostDischargeRequestWithBody(server, "application/json", bodyReader)
}

// NewPostDischargeRequestWithBody generates requests for PostDischarge with any type of body
func NewPostDischargeRequestWithBody(server string, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/discharge")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	return req, nil
}

// NewGetIdentitiesRequest generates requests for GetIdentities
func NewGetIdentitiesRequest(server string) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/identities")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewPostLoginRequest calls the generic PostLogin builder with application/json body
func NewPostLoginRequest(server string, body PostLoginJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewPostLoginRequestWithBody(server, "application/json", bodyReader)
}

// NewPostLoginRequestWithBody generates requests for PostLogin with any type of body
func NewPostLoginRequestWithBody(server string, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/login")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	return req, nil
}

// NewPostValidateTokenRequest calls the generic PostValidateToken builder with application/json body
func NewPostValidateTokenRequest(server string, body PostValidateTokenJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewPostValidateTokenRequestWithBody(server, "application/json", bodyReader)
}

// NewPostValidateTokenRequestWithBody generates requests for PostValidateToken with any type of body
func NewPostValidateTokenRequestWithBody(server string, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/validate-token")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	return req, nil
}

func (c *Client) applyEditors(ctx context.Context, req *http.Request, additionalEditors []RequestEditorFn) error {
	for _, r := range c.RequestEditors {
		if err := r(ctx, req); err != nil {
			return err
		}
	}
	for _, r := range additionalEditors {
		if err := r(ctx, req); err != nil {
			return err
		}
	}
	return nil
}

// ClientWithResponses builds on ClientInterface to offer response payloads
type ClientWithResponses struct {
	ClientInterface
}

// NewClientWithResponses creates a new ClientWithResponses, which wraps
// Client with return type handling
func NewClientWithResponses(server string, opts ...ClientOption) (*ClientWithResponses, error) {
	client, err := NewClient(server, opts...)
	if err != nil {
		return nil, err
	}
	return &ClientWithResponses{client}, nil
}

// WithBaseURL overrides the baseURL.
func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) error {
		newBaseURL, err := url.Parse(baseURL)
		if err != nil {
			return err
		}
		c.Server = newBaseURL.String()
		return nil
	}
}

// ClientWithResponsesInterface is the interface specification for the client with responses above.
type ClientWithResponsesInterface interface {
	// PostDischargeWithBodyWithResponse request with any body
	PostDischargeWithBodyWithResponse(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*PostDischargeResponse, error)

	PostDischargeWithResponse(ctx context.Context, body PostDischargeJSONRequestBody, reqEditors ...RequestEditorFn) (*PostDischargeResponse, error)

	// GetIdentitiesWithResponse request
	GetIdentitiesWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*GetIdentitiesResponse, error)

	// PostLoginWithBodyWithResponse request with any body
	PostLoginWithBodyWithResponse(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*PostLoginResponse, error)

	PostLoginWithResponse(ctx context.Context, body PostLoginJSONRequestBody, reqEditors ...RequestEditorFn) (*PostLoginResponse, error)

	// PostValidateTokenWithBodyWithResponse request with any body
	PostValidateTokenWithBodyWithResponse(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*PostValidateTokenResponse, error)

	PostValidateTokenWithResponse(ctx context.Context, body PostValidateTokenJSONRequestBody, reqEditors ...RequestEditorFn) (*PostValidateTokenResponse, error)
}

type PostDischargeResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *DischargeMacaroonResponse
	JSONDefault  *ErrorResponse
}

// Status returns HTTPResponse.Status
func (r PostDischargeResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r PostDischargeResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type GetIdentitiesResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSONDefault  *IdentitiesResponse
}

// Status returns HTTPResponse.Status
func (r GetIdentitiesResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r GetIdentitiesResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type PostLoginResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *LoginResponse
	JSONDefault  *ErrorResponse
}

// Status returns HTTPResponse.Status
func (r PostLoginResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r PostLoginResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type PostValidateTokenResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *ValidateTokenResponse
	JSONDefault  *ErrorResponse
}

// Status returns HTTPResponse.Status
func (r PostValidateTokenResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r PostValidateTokenResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

// PostDischargeWithBodyWithResponse request with arbitrary body returning *PostDischargeResponse
func (c *ClientWithResponses) PostDischargeWithBodyWithResponse(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*PostDischargeResponse, error) {
	rsp, err := c.PostDischargeWithBody(ctx, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParsePostDischargeResponse(rsp)
}

func (c *ClientWithResponses) PostDischargeWithResponse(ctx context.Context, body PostDischargeJSONRequestBody, reqEditors ...RequestEditorFn) (*PostDischargeResponse, error) {
	rsp, err := c.PostDischarge(ctx, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParsePostDischargeResponse(rsp)
}

// GetIdentitiesWithResponse request returning *GetIdentitiesResponse
func (c *ClientWithResponses) GetIdentitiesWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*GetIdentitiesResponse, error) {
	rsp, err := c.GetIdentities(ctx, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseGetIdentitiesResponse(rsp)
}

// PostLoginWithBodyWithResponse request with arbitrary body returning *PostLoginResponse
func (c *ClientWithResponses) PostLoginWithBodyWithResponse(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*PostLoginResponse, error) {
	rsp, err := c.PostLoginWithBody(ctx, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParsePostLoginResponse(rsp)
}

func (c *ClientWithResponses) PostLoginWithResponse(ctx context.Context, body PostLoginJSONRequestBody, reqEditors ...RequestEditorFn) (*PostLoginResponse, error) {
	rsp, err := c.PostLogin(ctx, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParsePostLoginResponse(rsp)
}

// PostValidateTokenWithBodyWithResponse request with arbitrary body returning *PostValidateTokenResponse
func (c *ClientWithResponses) PostValidateTokenWithBodyWithResponse(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*PostValidateTokenResponse, error) {
	rsp, err := c.PostValidateTokenWithBody(ctx, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParsePostValidateTokenResponse(rsp)
}

func (c *ClientWithResponses) PostValidateTokenWithResponse(ctx context.Context, body PostValidateTokenJSONRequestBody, reqEditors ...RequestEditorFn) (*PostValidateTokenResponse, error) {
	rsp, err := c.PostValidateToken(ctx, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParsePostValidateTokenResponse(rsp)
}

// ParsePostDischargeResponse parses an HTTP response from a PostDischargeWithResponse call
func ParsePostDischargeResponse(rsp *http.Response) (*PostDischargeResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &PostDischargeResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest DischargeMacaroonResponse
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && true:
		var dest ErrorResponse
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSONDefault = &dest

	}

	return response, nil
}

// ParseGetIdentitiesResponse parses an HTTP response from a GetIdentitiesWithResponse call
func ParseGetIdentitiesResponse(rsp *http.Response) (*GetIdentitiesResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &GetIdentitiesResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && true:
		var dest IdentitiesResponse
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSONDefault = &dest

	}

	return response, nil
}

// ParsePostLoginResponse parses an HTTP response from a PostLoginWithResponse call
func ParsePostLoginResponse(rsp *http.Response) (*PostLoginResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &PostLoginResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest LoginResponse
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && true:
		var dest ErrorResponse
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSONDefault = &dest

	}

	return response, nil
}

// ParsePostValidateTokenResponse parses an HTTP response from a PostValidateTokenWithResponse call
func ParsePostValidateTokenResponse(rsp *http.Response) (*PostValidateTokenResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &PostValidateTokenResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest ValidateTokenResponse
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && true:
		var dest ErrorResponse
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSONDefault = &dest

	}

	return response, nil
}

// ServerInterface represents all server handlers.
type ServerInterface interface {

	// (POST /discharge)
	PostDischarge(w http.ResponseWriter, r *http.Request)

	// (GET /identities)
	GetIdentities(w http.ResponseWriter, r *http.Request)

	// (POST /login)
	PostLogin(w http.ResponseWriter, r *http.Request)

	// (POST /validate-token)
	PostValidateToken(w http.ResponseWriter, r *http.Request)
}

// ServerInterfaceWrapper converts contexts to parameters.
type ServerInterfaceWrapper struct {
	Handler            ServerInterface
	HandlerMiddlewares []MiddlewareFunc
	ErrorHandlerFunc   func(w http.ResponseWriter, r *http.Request, err error)
}

type MiddlewareFunc func(http.Handler) http.Handler

// PostDischarge operation middleware
func (siw *ServerInterfaceWrapper) PostDischarge(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

	ctx = context.WithValue(ctx, BearerHTTPAuthenticationScopes, []string{})

	r = r.WithContext(ctx)

	handler := http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.PostDischarge(w, r)
	}))

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r)
}

// GetIdentities operation middleware
func (siw *ServerInterfaceWrapper) GetIdentities(w http.ResponseWriter, r *http.Request) {

	handler := http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.GetIdentities(w, r)
	}))

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r)
}

// PostLogin operation middleware
func (siw *ServerInterfaceWrapper) PostLogin(w http.ResponseWriter, r *http.Request) {

	handler := http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.PostLogin(w, r)
	}))

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r)
}

// PostValidateToken operation middleware
func (siw *ServerInterfaceWrapper) PostValidateToken(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

	ctx = context.WithValue(ctx, BearerHTTPAuthenticationScopes, []string{})

	r = r.WithContext(ctx)

	handler := http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.PostValidateToken(w, r)
	}))

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r)
}

type UnescapedCookieParamError struct {
	ParamName string
	Err       error
}

func (e *UnescapedCookieParamError) Error() string {
	return fmt.Sprintf("error unescaping cookie parameter '%s'", e.ParamName)
}

func (e *UnescapedCookieParamError) Unwrap() error {
	return e.Err
}

type UnmarshalingParamError struct {
	ParamName string
	Err       error
}

func (e *UnmarshalingParamError) Error() string {
	return fmt.Sprintf("Error unmarshaling parameter %s as JSON: %s", e.ParamName, e.Err.Error())
}

func (e *UnmarshalingParamError) Unwrap() error {
	return e.Err
}

type RequiredParamError struct {
	ParamName string
}

func (e *RequiredParamError) Error() string {
	return fmt.Sprintf("Query argument %s is required, but not found", e.ParamName)
}

type RequiredHeaderError struct {
	ParamName string
	Err       error
}

func (e *RequiredHeaderError) Error() string {
	return fmt.Sprintf("Header parameter %s is required, but not found", e.ParamName)
}

func (e *RequiredHeaderError) Unwrap() error {
	return e.Err
}

type InvalidParamFormatError struct {
	ParamName string
	Err       error
}

func (e *InvalidParamFormatError) Error() string {
	return fmt.Sprintf("Invalid format for parameter %s: %s", e.ParamName, e.Err.Error())
}

func (e *InvalidParamFormatError) Unwrap() error {
	return e.Err
}

type TooManyValuesForParamError struct {
	ParamName string
	Count     int
}

func (e *TooManyValuesForParamError) Error() string {
	return fmt.Sprintf("Expected one value for %s, got %d", e.ParamName, e.Count)
}

// Handler creates http.Handler with routing matching OpenAPI spec.
func Handler(si ServerInterface) http.Handler {
	return HandlerWithOptions(si, StdHTTPServerOptions{})
}

// ServeMux is an abstraction of http.ServeMux.
type ServeMux interface {
	HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request))
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}

type StdHTTPServerOptions struct {
	BaseURL          string
	BaseRouter       ServeMux
	Middlewares      []MiddlewareFunc
	ErrorHandlerFunc func(w http.ResponseWriter, r *http.Request, err error)
}

// HandlerFromMux creates http.Handler with routing matching OpenAPI spec based on the provided mux.
func HandlerFromMux(si ServerInterface, m ServeMux) http.Handler {
	return HandlerWithOptions(si, StdHTTPServerOptions{
		BaseRouter: m,
	})
}

func HandlerFromMuxWithBaseURL(si ServerInterface, m ServeMux, baseURL string) http.Handler {
	return HandlerWithOptions(si, StdHTTPServerOptions{
		BaseURL:    baseURL,
		BaseRouter: m,
	})
}

// HandlerWithOptions creates http.Handler with additional options
func HandlerWithOptions(si ServerInterface, options StdHTTPServerOptions) http.Handler {
	m := options.BaseRouter

	if m == nil {
		m = http.NewServeMux()
	}
	if options.ErrorHandlerFunc == nil {
		options.ErrorHandlerFunc = func(w http.ResponseWriter, r *http.Request, err error) {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	}

	wrapper := ServerInterfaceWrapper{
		Handler:            si,
		HandlerMiddlewares: options.Middlewares,
		ErrorHandlerFunc:   options.ErrorHandlerFunc,
	}

	m.HandleFunc("POST "+options.BaseURL+"/discharge", wrapper.PostDischarge)
	m.HandleFunc("GET "+options.BaseURL+"/identities", wrapper.GetIdentities)
	m.HandleFunc("POST "+options.BaseURL+"/login", wrapper.PostLogin)
	m.HandleFunc("POST "+options.BaseURL+"/validate-token", wrapper.PostValidateToken)

	return m
}
