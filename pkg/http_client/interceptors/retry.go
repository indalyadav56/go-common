package interceptors

import "net/http"

type RetryInterceptor struct {
	next http.RoundTripper
}

func NewRetryInterceptor(next http.RoundTripper) *RetryInterceptor {
	if next == nil {
		next = http.DefaultTransport
	}

	return &RetryInterceptor{
		next: next,
	}
}

func (r *RetryInterceptor) RoundTrip(req *http.Request) (*http.Response, error) {
	return r.next.RoundTrip(req)
}
