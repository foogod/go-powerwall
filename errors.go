package powerwall

import (
	"fmt"
	"net/url"
)

// ApiError indicates that something unexpected occurred with the HTTP API
// call.  This usually occurs when the endpoint returns an unexpected status
// code.
type ApiError struct {
	URL        url.URL
	StatusCode int
	Body       []byte
}

func (e ApiError) Error() string {
	return fmt.Sprintf("API call to %s returned unexpected status code %d (%#v)", e.URL.String(), e.StatusCode, string(e.Body))
}

// AuthFailure is returned when the client was unable to perform a request
// because it was not able to login using the provided email and password.
type AuthFailure struct {
	URL       url.URL
	ErrorText string
	Message   string
}

func (e AuthFailure) Error() string {
	return fmt.Sprintf("Authentication Failed: %s (%s)", e.ErrorText, e.Message)
}
