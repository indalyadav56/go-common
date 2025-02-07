package interceptors

import "net/http"

type AuthInterceptor struct {
	Next http.RoundTripper
}

func NewAuthInterceptor(next http.RoundTripper) *AuthInterceptor {
	if next == nil {
		next = http.DefaultTransport
	}
	return &AuthInterceptor{Next: next}
}

func (a *AuthInterceptor) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "Bearer <token>")
	return a.Next.RoundTrip(req)
}
