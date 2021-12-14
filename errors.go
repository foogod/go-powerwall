package powerwall

import (
	"fmt"
	"net/url"
)

type ApiError struct {
	URL        url.URL
	StatusCode int
	Body       []byte
}

func (e ApiError) Error() string {
	return fmt.Sprintf("API call to %s returned unexpected status code %d (%#v)", e.URL.String(), e.StatusCode, string(e.Body))
}

type AuthFailure struct {
	URL       url.URL
	ErrorText string
	Message   string
}

func (e AuthFailure) Error() string {
	return fmt.Sprintf("Authentication Failed: %s (%s)", e.ErrorText, e.Message)
}
