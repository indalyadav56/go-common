package interceptors

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
)

// Interceptor defines the interface for HTTP client interceptors
type Interceptor interface {
	RoundTrip(req *http.Request) (*http.Response, error)
}

// Logger interface defines the logging behavior
type Logger interface {
	Debug(msg string, fields ...interface{})
	Info(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
}

// LoggingInterceptor implements http.RoundTripper and provides detailed request/response logging
type LoggingInterceptor struct {
	Next           http.RoundTripper
	Logger         Logger
	LogRequestBody bool // Option to enable/disable request body logging
	LogHeaders     bool // Option to enable/disable headers logging
	MaxBodySize    int  // Maximum body size to log (in bytes)
}

// LoggingOptions contains configuration for the logging interceptor
type LoggingOptions struct {
	LogRequestBody bool
	LogHeaders     bool
	MaxBodySize    int
}

// NewLoggingInterceptor creates a new logging interceptor with the given options
func NewLoggingInterceptor(next http.RoundTripper, logger Logger, opts LoggingOptions) *LoggingInterceptor {
	if next == nil {
		next = http.DefaultTransport
	}
	if logger == nil {
		panic("logger cannot be nil")
	}
	return &LoggingInterceptor{
		Next:           next,
		Logger:         logger,
		LogRequestBody: opts.LogRequestBody,
		LogHeaders:     opts.LogHeaders,
		MaxBodySize:    opts.MaxBodySize,
	}
}

// RoundTrip implements http.RoundTripper interface
func (l *LoggingInterceptor) RoundTrip(req *http.Request) (*http.Response, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	// Generate request ID if not present
	reqID := req.Header.Get("X-Request-ID")
	if reqID == "" {
		reqID = uuid.New().String()
		req.Header.Set("X-Request-ID", reqID)
	}

	// Create context with timeout if not present
	ctx := req.Context()
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
		req = req.WithContext(ctx)
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

	l.Logger.Info("outgoing request", fields)

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

	l.Logger.Info("received response", respFields)
	return resp, nil
}

// readBody reads the body while preserving it for future reads
func (l *LoggingInterceptor) readBody(body io.ReadCloser) ([]byte, error) {
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

// prettyJSON formats JSON data for logging
func prettyJSON(data interface{}) string {
	formatted, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Sprintf("[error formatting JSON: %v]", err)
	}
	return string(formatted)
}

// StandardLogger implements the Logger interface using the standard log package
type StandardLogger struct {
	logger *log.Logger
}

// NewStandardLogger creates a new StandardLogger
func NewStandardLogger() *StandardLogger {
	return &StandardLogger{
		logger: log.New(os.Stdout, "[HTTP] ", log.LstdFlags),
	}
}

func (l *StandardLogger) Debug(msg string, fields ...interface{}) {
	l.log("DEBUG", msg, fields...)
}

func (l *StandardLogger) Info(msg string, fields ...interface{}) {
	l.log("INFO", msg, fields...)
}

func (l *StandardLogger) Error(msg string, fields ...interface{}) {
	l.log("ERROR", msg, fields...)
}

func (l *StandardLogger) log(level, msg string, fields ...interface{}) {
	// Convert fields to a map
	fieldsMap := make(map[string]interface{})
	for i := 0; i < len(fields); i += 2 {
		if i+1 < len(fields) {
			fieldsMap[fmt.Sprint(fields[i])] = fields[i+1]
		}
	}

	// Add message to fields
	fieldsMap["msg"] = msg
	fieldsMap["level"] = level

	// Marshal to JSON
	output, err := json.Marshal(fieldsMap)
	if err != nil {
		l.logger.Printf("[%s] %s (error marshaling fields: %v)", level, msg, err)
		return
	}

	l.logger.Println(string(output))
}
