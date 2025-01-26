package http_client

import (
	"net/http"
	"time"
)

// Config holds the configuration for the HTTP client
type Config struct {
	// BaseURL for all requests
	BaseURL string

	// Timeout for requests
	Timeout time.Duration

	// GlobalHeaders to be added to all requests
	GlobalHeaders map[string]string

	// Interceptor for the client
	Interceptor http.RoundTripper

	// MaxIdleConns controls the maximum number of idle (keep-alive) connections across all hosts
	MaxIdleConns int

	// MaxIdleConnsPerHost controls the maximum idle (keep-alive) connections to keep per-host
	MaxIdleConnsPerHost int

	// MaxConnsPerHost optionally limits the total number of connections per host
	MaxConnsPerHost int

	// IdleConnTimeout is the maximum amount of time an idle connection will remain idle before closing
	IdleConnTimeout time.Duration

	// TLSHandshakeTimeout specifies the maximum amount of time waiting to wait for a TLS handshake
	TLSHandshakeTimeout time.Duration

	// DisableKeepAlives, if true, prevents re-use of TCP connections
	DisableKeepAlives bool

	// DisableCompression, if true, prevents the Transport from requesting compression
	DisableCompression bool

	// ResponseHeaderTimeout, if non-zero, specifies the amount of time to wait for a server's response headers
	ResponseHeaderTimeout time.Duration
}

// Option defines a function that can modify the Config
type Option func(*Config)

// defaultConfig returns the default configuration
func defaultConfig(config ...Config) Config {
	cfg := Config{
		Timeout:               30 * time.Second,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   10,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
	}

	if len(config) > 0 {
		cfg = config[0]
	}

	return cfg
}

// WithBaseURL sets the base URL for the client
func WithBaseURL(url string) Option {
	return func(c *Config) {
		c.BaseURL = url
	}
}

// WithTimeout sets the timeout for requests
func WithTimeout(timeout time.Duration) Option {
	return func(c *Config) {
		c.Timeout = timeout
	}
}

// WithGlobalHeaders sets global headers for all requests
func WithGlobalHeaders(headers map[string]string) Option {
	return func(c *Config) {
		c.GlobalHeaders = headers
	}
}

// WithMaxIdleConns sets the maximum number of idle connections
func WithMaxIdleConns(n int) Option {
	return func(c *Config) {
		c.MaxIdleConns = n
	}
}

// WithMaxIdleConnsPerHost sets the maximum number of idle connections per host
func WithMaxIdleConnsPerHost(n int) Option {
	return func(c *Config) {
		c.MaxIdleConnsPerHost = n
	}
}

// WithDisableKeepAlives sets whether to disable keep-alives
func WithDisableKeepAlives(disable bool) Option {
	return func(c *Config) {
		c.DisableKeepAlives = disable
	}
}

// WithDisableCompression sets whether to disable compression
func WithDisableCompression(disable bool) Option {
	return func(c *Config) {
		c.DisableCompression = disable
	}
}
