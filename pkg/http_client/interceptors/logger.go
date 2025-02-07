package interceptors

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
)

type Interceptor interface {
	RoundTrip(req *http.Request) (*http.Response, error)
}

type Logger interface {
	Debug(msg string, fields ...interface{})
	Info(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
}

type LoggerInterceptor struct {
	Next           http.RoundTripper
	Logger         Logger
	LogRequestBody bool
	LogHeaders     bool
	MaxBodySize    int
}

type LoggingOptions struct {
	LogRequestBody bool
	LogHeaders     bool
	MaxBodySize    int
}

func NewLoggerInterceptor(next http.RoundTripper, logger Logger, opts ...*LoggingOptions) *LoggerInterceptor {
	if next == nil {
		next = http.DefaultTransport
	}

	if logger == nil {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))
	}

	if len(opts) < 1 {
		opts = []*LoggingOptions{
			{
				LogRequestBody: true,
				LogHeaders:     true,
				MaxBodySize:    4096,
			},
		}
	}

	return &LoggerInterceptor{
		Next:           next,
		Logger:         logger,
		LogRequestBody: opts[0].LogRequestBody,
		LogHeaders:     opts[0].LogHeaders,
		MaxBodySize:    opts[0].MaxBodySize,
	}
}

// Convert map to key-value pairs for slog
func flattenFields(fields map[string]interface{}) []interface{} {
	flattened := make([]interface{}, 0, len(fields)*2)
	for k, v := range fields {
		flattened = append(flattened, k, v)
	}
	return flattened
}

func (l *LoggerInterceptor) RoundTrip(req *http.Request) (*http.Response, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	// Generate request ID if not present
	reqID := req.Header.Get("X-Request-ID")
	if reqID == "" {
		reqID = uuid.New().String()
		req.Header.Set("X-Request-ID", reqID)
	}

	start := time.Now()
	fields := map[string]interface{}{
		"request_id": reqID,
		"method":     req.Method,
		"url":        req.URL.String(),
	}

	if l.LogHeaders {
		fields["headers"] = headerToMap(req.Header)
	}

	if l.LogRequestBody && req.Body != nil {
		body, err := l.readBody(req.Body)
		if err != nil {
			l.Logger.Error("failed to read request body", "error", err)
		} else {
			req.Body = io.NopCloser(bytes.NewBuffer(body))
			fields["body"] = truncateBody(body, l.MaxBodySize)
		}
	}

	l.Logger.Info("outgoing request", flattenFields(fields)...)

	// Make the actual request
	resp, err := l.Next.RoundTrip(req)
	duration := time.Since(start)

	if err != nil {
		l.Logger.Error("request failed",
			"request_id", reqID,
			"duration", duration.String(),
			"error", err.Error(),
		)
		return nil, fmt.Errorf("request failed: %w", err)
	}

	// Log response
	respFields := map[string]interface{}{
		"request_id":   reqID,
		"status_code":  resp.StatusCode,
		"duration":     duration.String(),
		"content_type": resp.Header.Get("Content-Type"),
		"content_size": resp.ContentLength,
	}

	if l.LogHeaders {
		respFields["headers"] = headerToMap(resp.Header)
	}

	if l.LogRequestBody && resp.Body != nil {
		body, err := l.readBody(resp.Body)
		if err != nil {
			l.Logger.Error("failed to read response body", "error", err)
		} else {
			resp.Body = io.NopCloser(bytes.NewBuffer(body))
			respFields["body"] = truncateBody(body, l.MaxBodySize)
		}
	}

	l.Logger.Info("received response", flattenFields(fields)...)
	return resp, nil
}

// readBody reads the body while preserving it for future reads
func (l *LoggerInterceptor) readBody(body io.ReadCloser) ([]byte, error) {
	if body == nil {
		return nil, nil
	}
	data, err := io.ReadAll(body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body: %w", err)
	}
	return data, nil
}

// headerToMap converts http.Header to a map for structured logging
func headerToMap(header http.Header) map[string][]string {
	result := make(map[string][]string, len(header))
	for k, v := range header {
		result[k] = v
	}
	return result
}

// truncateBody truncates the body if it exceeds maxSize
func truncateBody(body []byte, maxSize int) string {
	if maxSize <= 0 || len(body) <= maxSize {
		return string(body)
	}
	return fmt.Sprintf("%s... [truncated %d bytes]", string(body[:maxSize]), len(body)-maxSize)
}
