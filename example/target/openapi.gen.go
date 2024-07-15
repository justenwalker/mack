//go:build go1.22

// Package targetsvc provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen/v2 version v2.2.0 DO NOT EDIT.
package target

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/oapi-codegen/runtime"
)

// ErrorResponseBody defines model for ErrorResponseBody.
type ErrorResponseBody struct {
	// Code error code
	Code int `json:"code"`

	// Error error message
	Error string `json:"error"`
}

// MacaroonResponseBody defines model for MacaroonResponseBody.
type MacaroonResponseBody struct {
	// ExpiresIn number of seconds for which this macaroon is valid after this response was generated
	ExpiresIn int64 `json:"expires_in"`

	// Macaroon A macaroon discharging the provided caveat ID
	Macaroon string `json:"macaroon"`
}

// OperationRequestBody defines model for OperationRequestBody.
type OperationRequestBody struct {
	// Arguments argument list
	Arguments *[]string `json:"arguments,omitempty"`

	// Operation operation id
	Operation string `json:"operation"`
}

// OperationResponseBody defines model for OperationResponseBody.
type OperationResponseBody = map[string]interface{}

// ErrorResponse defines model for ErrorResponse.
type ErrorResponse = ErrorResponseBody

// MacaroonResponse defines model for MacaroonResponse.
type MacaroonResponse = MacaroonResponseBody

// OperationResponse defines model for OperationResponse.
type OperationResponse = OperationResponseBody

// OperationRequest defines model for OperationRequest.
type OperationRequest = OperationRequestBody

// GetMacaroonRequestParams defines parameters for GetMacaroonRequest.
type GetMacaroonRequestParams struct {
	Org *string `form:"org,omitempty" json:"org,omitempty"`
	App *string `form:"app,omitempty" json:"app,omitempty"`
}

// PostOrgAppDoJSONRequestBody defines body for PostOrgAppDo for application/json ContentType.
type PostOrgAppDoJSONRequestBody = OperationRequestBody

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
	// customized settings, such as certificate stacks.
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
	// GetMacaroonRequest request
	GetMacaroonRequest(ctx context.Context, params *GetMacaroonRequestParams, reqEditors ...RequestEditorFn) (*http.Response, error)

	// PostOrgAppDoWithBody request with any body
	PostOrgAppDoWithBody(ctx context.Context, org string, app string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	PostOrgAppDo(ctx context.Context, org string, app string, body PostOrgAppDoJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)
}

func (c *Client) GetMacaroonRequest(ctx context.Context, params *GetMacaroonRequestParams, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewGetMacaroonRequestRequest(c.Server, params)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) PostOrgAppDoWithBody(ctx context.Context, org string, app string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewPostOrgAppDoRequestWithBody(c.Server, org, app, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) PostOrgAppDo(ctx context.Context, org string, app string, body PostOrgAppDoJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewPostOrgAppDoRequest(c.Server, org, app, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

// NewGetMacaroonRequestRequest generates requests for GetMacaroonRequest
func NewGetMacaroonRequestRequest(server string, params *GetMacaroonRequestParams) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/macaroon-request")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	if params != nil {
		queryValues := queryURL.Query()

		if params.Org != nil {

			if queryFrag, err := runtime.StyleParamWithLocation("form", true, "org", runtime.ParamLocationQuery, *params.Org); err != nil {
				return nil, err
			} else if parsed, err := url.ParseQuery(queryFrag); err != nil {
				return nil, err
			} else {
				for k, v := range parsed {
					for _, v2 := range v {
						queryValues.Add(k, v2)
					}
				}
			}

		}

		if params.App != nil {

			if queryFrag, err := runtime.StyleParamWithLocation("form", true, "app", runtime.ParamLocationQuery, *params.App); err != nil {
				return nil, err
			} else if parsed, err := url.ParseQuery(queryFrag); err != nil {
				return nil, err
			} else {
				for k, v := range parsed {
					for _, v2 := range v {
						queryValues.Add(k, v2)
					}
				}
			}

		}

		queryURL.RawQuery = queryValues.Encode()
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewPostOrgAppDoRequest calls the generic PostOrgAppDo builder with application/json body
func NewPostOrgAppDoRequest(server string, org string, app string, body PostOrgAppDoJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewPostOrgAppDoRequestWithBody(server, org, app, "application/json", bodyReader)
}

// NewPostOrgAppDoRequestWithBody generates requests for PostOrgAppDo with any type of body
func NewPostOrgAppDoRequestWithBody(server string, org string, app string, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "org", runtime.ParamLocationPath, org)
	if err != nil {
		return nil, err
	}

	var pathParam1 string

	pathParam1, err = runtime.StyleParamWithLocation("simple", false, "app", runtime.ParamLocationPath, app)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/%s/%s/do", pathParam0, pathParam1)
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
	// GetMacaroonRequestWithResponse request
	GetMacaroonRequestWithResponse(ctx context.Context, params *GetMacaroonRequestParams, reqEditors ...RequestEditorFn) (*GetMacaroonRequestResponse, error)

	// PostOrgAppDoWithBodyWithResponse request with any body
	PostOrgAppDoWithBodyWithResponse(ctx context.Context, org string, app string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*PostOrgAppDoResponse, error)

	PostOrgAppDoWithResponse(ctx context.Context, org string, app string, body PostOrgAppDoJSONRequestBody, reqEditors ...RequestEditorFn) (*PostOrgAppDoResponse, error)
}

type GetMacaroonRequestResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *MacaroonResponse
	JSONDefault  *ErrorResponse
}

// Status returns HTTPResponse.Status
func (r GetMacaroonRequestResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r GetMacaroonRequestResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type PostOrgAppDoResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *OperationResponse
	JSONDefault  *ErrorResponse
}

// Status returns HTTPResponse.Status
func (r PostOrgAppDoResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r PostOrgAppDoResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

// GetMacaroonRequestWithResponse request returning *GetMacaroonRequestResponse
func (c *ClientWithResponses) GetMacaroonRequestWithResponse(ctx context.Context, params *GetMacaroonRequestParams, reqEditors ...RequestEditorFn) (*GetMacaroonRequestResponse, error) {
	rsp, err := c.GetMacaroonRequest(ctx, params, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseGetMacaroonRequestResponse(rsp)
}

// PostOrgAppDoWithBodyWithResponse request with arbitrary body returning *PostOrgAppDoResponse
func (c *ClientWithResponses) PostOrgAppDoWithBodyWithResponse(ctx context.Context, org string, app string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*PostOrgAppDoResponse, error) {
	rsp, err := c.PostOrgAppDoWithBody(ctx, org, app, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParsePostOrgAppDoResponse(rsp)
}

func (c *ClientWithResponses) PostOrgAppDoWithResponse(ctx context.Context, org string, app string, body PostOrgAppDoJSONRequestBody, reqEditors ...RequestEditorFn) (*PostOrgAppDoResponse, error) {
	rsp, err := c.PostOrgAppDo(ctx, org, app, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParsePostOrgAppDoResponse(rsp)
}

// ParseGetMacaroonRequestResponse parses an HTTP response from a GetMacaroonRequestWithResponse call
func ParseGetMacaroonRequestResponse(rsp *http.Response) (*GetMacaroonRequestResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &GetMacaroonRequestResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest MacaroonResponse
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

// ParsePostOrgAppDoResponse parses an HTTP response from a PostOrgAppDoWithResponse call
func ParsePostOrgAppDoResponse(rsp *http.Response) (*PostOrgAppDoResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &PostOrgAppDoResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest OperationResponse
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

	// (GET /macaroon-request)
	GetMacaroonRequest(w http.ResponseWriter, r *http.Request, params GetMacaroonRequestParams)

	// (POST /{org}/{app}/do)
	PostOrgAppDo(w http.ResponseWriter, r *http.Request, org string, app string)
}

// ServerInterfaceWrapper converts contexts to parameters.
type ServerInterfaceWrapper struct {
	Handler            ServerInterface
	HandlerMiddlewares []MiddlewareFunc
	ErrorHandlerFunc   func(w http.ResponseWriter, r *http.Request, err error)
}

type MiddlewareFunc func(http.Handler) http.Handler

// GetMacaroonRequest operation middleware
func (siw *ServerInterfaceWrapper) GetMacaroonRequest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var err error

	// Parameter object where we will unmarshal all parameters from the context
	var params GetMacaroonRequestParams

	// ------------- Optional query parameter "org" -------------

	err = runtime.BindQueryParameter("form", true, false, "org", r.URL.Query(), &params.Org)
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "org", Err: err})
		return
	}

	// ------------- Optional query parameter "app" -------------

	err = runtime.BindQueryParameter("form", true, false, "app", r.URL.Query(), &params.App)
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "app", Err: err})
		return
	}

	handler := http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.GetMacaroonRequest(w, r, params)
	}))

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r.WithContext(ctx))
}

// PostOrgAppDo operation middleware
func (siw *ServerInterfaceWrapper) PostOrgAppDo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var err error

	// ------------- Path parameter "org" -------------
	var org string

	err = runtime.BindStyledParameterWithOptions("simple", "org", r.PathValue("org"), &org, runtime.BindStyledParameterOptions{ParamLocation: runtime.ParamLocationPath, Explode: false, Required: false})
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "org", Err: err})
		return
	}

	// ------------- Path parameter "app" -------------
	var app string

	err = runtime.BindStyledParameterWithOptions("simple", "app", r.PathValue("app"), &app, runtime.BindStyledParameterOptions{ParamLocation: runtime.ParamLocationPath, Explode: false, Required: false})
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "app", Err: err})
		return
	}

	handler := http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.PostOrgAppDo(w, r, org, app)
	}))

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r.WithContext(ctx))
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

type StdHTTPServerOptions struct {
	BaseURL          string
	BaseRouter       *http.ServeMux
	Middlewares      []MiddlewareFunc
	ErrorHandlerFunc func(w http.ResponseWriter, r *http.Request, err error)
}

// HandlerFromMux creates http.Handler with routing matching OpenAPI spec based on the provided mux.
func HandlerFromMux(si ServerInterface, m *http.ServeMux) http.Handler {
	return HandlerWithOptions(si, StdHTTPServerOptions{
		BaseRouter: m,
	})
}

func HandlerFromMuxWithBaseURL(si ServerInterface, m *http.ServeMux, baseURL string) http.Handler {
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

	m.HandleFunc("GET "+options.BaseURL+"/macaroon-request", wrapper.GetMacaroonRequest)
	m.HandleFunc("POST "+options.BaseURL+"/{org}/{app}/do", wrapper.PostOrgAppDo)

	return m
}