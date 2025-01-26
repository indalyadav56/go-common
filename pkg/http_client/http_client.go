package http_client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
)

type Client interface {
	Get(ctx context.Context, endpoint string) RequestBuilder
	Post(ctx context.Context, endpoint string) RequestBuilder
	Put(ctx context.Context, endpoint string) RequestBuilder
	Patch(ctx context.Context, endpoint string) RequestBuilder
	Delete(ctx context.Context, endpoint string) RequestBuilder

	// Authentication methods
	SetBearerToken(token string) Client
	WithBasicAuth(username, password string) Client
}

type RequestBuilder interface {
	SetHeader(key, value string) RequestBuilder
	SetHeaders(headers map[string]string) RequestBuilder
	SetBody(body interface{}) RequestBuilder
	SetQueryParam(key, value string) RequestBuilder
	SetQueryParams(params map[string]string) RequestBuilder
	AddFile(fieldName, filePath string) RequestBuilder
	OnSuccess(fn func(*Response)) RequestBuilder
	OnError(fn func(*RequestError)) RequestBuilder
	SetError(v interface{}) RequestBuilder
	Into(v interface{}) error
	Result() (*Response, error)
}

type BatchRequest interface {
	Add(rb RequestBuilder) BatchRequest
	Execute(ctx context.Context) ([]*Response, []error)
}

type RequestPool interface {
	Submit(rb RequestBuilder) <-chan Result
	Wait()
}

type Result struct {
	Response *Response
	Error    error
}

type client struct {
	httpClient    *http.Client
	baseURL       string
	globalHeaders map[string]string
	interceptor   http.RoundTripper
	bearerToken   string
	basicAuth     struct {
		Username string
		Password string
	}
}

type request struct {
	client         *client
	method         string
	endpoint       string
	ctx            context.Context
	headers        map[string]string
	body           interface{}
	files          map[string]string
	queryParams    map[string]string
	successHandler func(*Response)
	errorHandler   func(*RequestError)
	errorType      interface{}
	result         interface{}
	executed       bool
	response       *Response
	err            error
}

func New(config ...Config) Client {
	cfg := defaultConfig(config...)

	transport := http.DefaultTransport

	if cfg.Interceptor != nil {
		transport = cfg.Interceptor
	}

	c := &client{
		httpClient: &http.Client{
			Timeout:   cfg.Timeout,
			Transport: transport,
		},
		baseURL:       cfg.BaseURL,
		globalHeaders: cfg.GlobalHeaders,
		interceptor:   cfg.Interceptor,
	}

	return c
}

func (c *client) Get(ctx context.Context, endpoint string) RequestBuilder {
	return &request{
		client:      c,
		method:      http.MethodGet,
		endpoint:    endpoint,
		ctx:         ctx,
		headers:     make(map[string]string),
		queryParams: make(map[string]string),
		files:       make(map[string]string),
	}
}

func (c *client) Post(ctx context.Context, endpoint string) RequestBuilder {
	return &request{
		client:      c,
		method:      http.MethodPost,
		endpoint:    endpoint,
		ctx:         ctx,
		headers:     make(map[string]string),
		queryParams: make(map[string]string),
		files:       make(map[string]string),
	}
}

func (c *client) Put(ctx context.Context, endpoint string) RequestBuilder {
	return &request{
		client:      c,
		method:      http.MethodPut,
		endpoint:    endpoint,
		ctx:         ctx,
		headers:     make(map[string]string),
		queryParams: make(map[string]string),
		files:       make(map[string]string),
	}
}

func (c *client) Patch(ctx context.Context, endpoint string) RequestBuilder {
	return &request{
		client:      c,
		method:      http.MethodPatch,
		endpoint:    endpoint,
		ctx:         ctx,
		headers:     make(map[string]string),
		queryParams: make(map[string]string),
		files:       make(map[string]string),
	}
}

func (c *client) Delete(ctx context.Context, endpoint string) RequestBuilder {
	return &request{
		client:      c,
		method:      http.MethodDelete,
		endpoint:    endpoint,
		ctx:         ctx,
		headers:     make(map[string]string),
		queryParams: make(map[string]string),
		files:       make(map[string]string),
	}
}

// SetBearerToken sets the bearer token for authentication
func (c *client) SetBearerToken(token string) Client {
	c.bearerToken = token
	return c
}

// WithBasicAuth sets the basic auth credentials
func (c *client) WithBasicAuth(username, password string) Client {
	c.basicAuth.Username = username
	c.basicAuth.Password = password
	return c
}

// Request implementation
func (r *request) Result() (*Response, error) {
	if !r.executed {
		r.execute()
	}

	return r.response, r.err
}

func (r *request) Into(v interface{}) error {
	resp, err := r.Result()
	if err != nil {
		// If it's a RequestError and we have an error type set, try to unmarshal
		if reqErr, ok := err.(*RequestError); ok && r.errorType != nil {
			if unmarshalErr := json.Unmarshal(reqErr.Response, r.errorType); unmarshalErr == nil {
				// Add the unmarshaled error details to the error
				return fmt.Errorf("%w: %+v", err, r.errorType)
			}
		}
		return err
	}
	return json.Unmarshal(resp.Body, v)
}

func (r *request) SetError(v interface{}) RequestBuilder {
	r.errorType = v
	return r
}

// RequestBuilder implementation methods
func (r *request) SetHeader(key, value string) RequestBuilder {
	if r.headers == nil {
		r.headers = make(map[string]string)
	}
	r.headers[key] = value
	return r
}

func (r *request) SetHeaders(headers map[string]string) RequestBuilder {
	if r.headers == nil {
		r.headers = make(map[string]string)
	}
	for k, v := range headers {
		r.headers[k] = v
	}
	return r
}

func (r *request) SetBody(body interface{}) RequestBuilder {
	r.body = body
	return r
}

func (r *request) SetQueryParam(key, value string) RequestBuilder {
	if r.queryParams == nil {
		r.queryParams = make(map[string]string)
	}
	r.queryParams[key] = value
	return r
}

func (r *request) SetQueryParams(params map[string]string) RequestBuilder {
	if r.queryParams == nil {
		r.queryParams = make(map[string]string)
	}
	for k, v := range params {
		r.queryParams[k] = v
	}
	return r
}

func (r *request) OnSuccess(fn func(*Response)) RequestBuilder {
	r.successHandler = fn
	if r.executed && r.err == nil && r.response != nil {
		fn(r.response)
	}
	return r
}

func (r *request) OnError(fn func(*RequestError)) RequestBuilder {
	r.errorHandler = fn
	if r.executed && r.err != nil {
		if reqErr, ok := r.err.(*RequestError); ok {
			fn(reqErr)
		}
	}
	return r
}

func (r *request) AddFile(fieldName, filePath string) RequestBuilder {
	r.files[fieldName] = filePath
	return r
}

// Response type remains the same
type Response struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
}

// RequestError type remains the same
type RequestError struct {
	StatusCode int
	URL        string
	Method     string
	Response   []byte
	Err        error
}

func (e *RequestError) Error() string {
	return fmt.Sprintf("request failed: method=%s, url=%s, status=%d, error=%v",
		e.Method, e.URL, e.StatusCode, e.Err)
}

func (e *RequestError) Unwrap() error {
	return e.Err
}

func (r *request) execute() {
	if r.executed {
		return
	}

	// Prepare URL with query parameters
	resolvedURL, err := r.client.resolveURL(r.endpoint)
	if err != nil {
		r.err = fmt.Errorf("failed to resolve URL: %w", err)
		r.executed = true
		return
	}

	parsedURL, err := url.Parse(resolvedURL)
	if err != nil {
		r.err = fmt.Errorf("invalid URL: %w", err)
		r.executed = true
		return
	}

	if len(r.queryParams) > 0 {
		q := parsedURL.Query()
		for k, v := range r.queryParams {
			q.Set(k, v)
		}
		parsedURL.RawQuery = q.Encode()
	}

	// Prepare body
	var bodyReader io.Reader
	if r.body != nil {
		bodyBytes, err := r.prepareBody()
		if err != nil {
			r.err = fmt.Errorf("failed to prepare request body: %w", err)
			r.executed = true
			return
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	// Create request
	req, err := http.NewRequestWithContext(r.ctx, r.method, parsedURL.String(), bodyReader)
	if err != nil {
		r.err = fmt.Errorf("failed to create request: %w", err)
		r.executed = true
		return
	}

	// Add headers
	r.addHeaders(req)

	// Add authentication headers
	if r.client.bearerToken != "" {
		req.Header.Set("Authorization", "Bearer "+r.client.bearerToken)
	}
	if r.client.basicAuth.Username != "" && r.client.basicAuth.Password != "" {
		req.SetBasicAuth(r.client.basicAuth.Username, r.client.basicAuth.Password)
	}

	// Add files
	if len(r.files) > 0 {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		for fieldName, filePath := range r.files {
			part, err := writer.CreateFormFile(fieldName, filePath)
			if err != nil {
				r.err = fmt.Errorf("failed to create form file: %w", err)
				r.executed = true
				return
			}
			file, err := os.Open(filePath)
			if err != nil {
				r.err = fmt.Errorf("failed to open file: %w", err)
				r.executed = true
				return
			}
			defer file.Close()
			_, err = io.Copy(part, file)
			if err != nil {
				r.err = fmt.Errorf("failed to copy file: %w", err)
				r.executed = true
				return
			}
		}
		err = writer.Close()
		if err != nil {
			r.err = fmt.Errorf("failed to close writer: %w", err)
			r.executed = true
			return
		}
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.Body = io.NopCloser(body)
	}

	// Execute request
	resp, err := r.client.httpClient.Do(req)
	if err != nil {
		if r.ctx.Err() != nil {
			r.err = fmt.Errorf("request canceled or timed out: %w", r.ctx.Err())
		} else {
			r.err = fmt.Errorf("request failed: %w", err)
		}
		r.executed = true
		return
	}
	defer func() {
		if resp.Body != nil {
			_, _ = io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		r.err = fmt.Errorf("error reading response body: %w", err)
		r.executed = true
		return
	}

	if resp.StatusCode >= 400 {
		reqErr := &RequestError{
			StatusCode: resp.StatusCode,
			URL:        req.URL.String(),
			Method:     req.Method,
			Response:   body,
			Err:        fmt.Errorf("request failed with status code %d", resp.StatusCode),
		}

		// Try to unmarshal error response if error type is set
		if r.errorType != nil {
			if err := json.Unmarshal(body, r.errorType); err == nil {
				reqErr.Err = fmt.Errorf("request failed with status code %d: %+v", resp.StatusCode, r.errorType)
			}
		}

		r.err = reqErr
		r.executed = true
		return
	}

	r.response = &Response{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
		Body:       body,
	}

	// Try to unmarshal success response if result type is set
	if r.result != nil {
		if err := json.Unmarshal(body, r.result); err != nil {
			r.err = fmt.Errorf("failed to unmarshal response: %w", err)
			r.executed = true
			return
		}
	}

	r.executed = true
}

func (r *request) prepareBody() ([]byte, error) {
	if r.body == nil {
		return nil, nil
	}

	switch body := r.body.(type) {
	case []byte:
		return body, nil
	case string:
		return []byte(body), nil
	case io.Reader:
		return io.ReadAll(body)
	default:
		return json.Marshal(body)
	}
}

func (r *request) addHeaders(req *http.Request) {
	// Set default headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Add global headers
	for key, value := range r.client.globalHeaders {
		req.Header.Set(key, value)
	}

	// Add request-specific headers
	for key, value := range r.headers {
		req.Header.Set(key, value)
	}
}

func (h *client) resolveURL(endpoint string) (string, error) {
	if h.baseURL == "" {
		return endpoint, nil
	}

	resolvedURL, err := url.JoinPath(h.baseURL, endpoint)
	if err != nil {
		return "", fmt.Errorf("failed to resolve URL: %w", err)
	}
	return resolvedURL, nil
}
