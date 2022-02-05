// Functions for configuring the client object:
//
//   (*Client) FetchTLSCert()
//   (*Client) SetTLSCert(cert)
//
package powerwall

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var logFunc = func(v ...interface{}) {}

// SetLogFunc registers a callback function which can be used for debug logging
// of the powerwall library.  The provided function should accept arguments in
// the same format as Printf/Sprintf/etc.  Note that log lines passed to this
// function are *not* newline-terminated, so you will need to add newlines if
// you want to put them out directly to stdout/stderr, etc.
func SetLogFunc(f func(...interface{})) {
	logFunc = f
}

var errFunc = func(string, error) {}

// SetErrFunc registers a callback function which will be called with
// additional information when certain errors occur.  This can be useful if you
// don't want full debug logging, but still want to log additional information
// that might be helpful when troubleshooting, for example, API message format
// errors, etc.
func SetErrFunc(f func(string, error)) {
	errFunc = f
}

// Client represents a connection to a Tesla Energy Gateway (Powerwall controller).
type Client struct {
	gatewayAddress       string
	gatewayLoginEmail    string
	gatewayLoginPassword string
	httpClient           *http.Client
	token_ch             chan string
	auth_ch              chan *authMessage
	retryInterval        time.Duration
	retryTimeout         time.Duration
}

// NewClient creates a new Client object.  gatewayAddress should be the IP
// address or hostname of the Tesla Energy Gateway (TEG) device which is
// connected to the network.  gatewayLoginEmail and gatewayLoginPassword are
// the same as the email address and password which would be used for a
// "customer" login, when logging in via the web interface (Note: This is not
// necessarily the same password as used to login to Tesla's servers via the
// app, which can be different)
//
// For more information on logging into an Energy Gateway, see
// https://www.tesla.com/support/energy/powerwall/own/monitoring-from-home-network
func NewClient(gatewayAddress, gatewayLoginEmail, gatewayLoginPassword string, options ...func(c *Client)) *Client {
	httpClient := &http.Client{
		Transport: DefaultTransport(),
		Timeout:   time.Second * 2, // Timeout after 2 seconds
	}

	c := &Client{
		gatewayAddress:       gatewayAddress,
		gatewayLoginEmail:    gatewayLoginEmail,
		gatewayLoginPassword: gatewayLoginPassword,
		httpClient:           httpClient,
		token_ch:             make(chan string),
		auth_ch:              make(chan *authMessage),
	}

	for _, option := range options {
		if option != nil {
			option(c)
		}
	}

	go c.authManager()

	c.logf("New powerwall client created: gateway_address=%s email=%s", gatewayAddress, gatewayLoginEmail)
	return c
}

// WithHttpClient sets the HTTP client to use for all requests
func WithHttpClient(httpClient *http.Client) func(c *Client) {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// DefaultTransport returns an HTTP transport with required TLS config
func DefaultTransport() *http.Transport {
	return &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
			// The TEG apparently requires a valid SNI hostname
			// matching its cert, or it will just bomb out and
			// terminate the connection during TLS negotiation
			// (even if we're not checking the cert), so we
			// override the TLS ServerName here to use one of its
			// (hardcoded) stock names for all connections.
			ServerName: "powerwall",
		},
	}
}

func (c *Client) logf(format string, v ...interface{}) {
	logFunc(fmt.Sprintf("{%p} ", c) + fmt.Sprintf(format, v...))
}

func (c *Client) jsonError(api string, data []byte, err error) {
	msg := fmt.Sprintf("Error unmarshalling '%s' response %s", api, string(data))
	errFunc(msg, err)
}

// SetTLSCert sets the TLS certificate which should be used for validating the
// certificate presented when connecting to the gateway is correct.  You can
// obtain the current certificate in use by the gateway initially via
// `FetchTLSCert`, and then supply it via this function whenever creating a new
// connection to ensure the certificate matches the previously fetched one.
func (c *Client) SetTLSCert(cert *x509.Certificate) {
	certPool := x509.NewCertPool()
	certPool.AddCert(cert)

	c.httpClient.Transport.(*http.Transport).TLSClientConfig.InsecureSkipVerify = false
	c.httpClient.Transport.(*http.Transport).TLSClientConfig.RootCAs = certPool
	c.logf("Set TLS cert for validation: %s", cert.Subject)
}

// FetchTLSCert queries the gateway and returns a copy of the TLS certificate
// it is currently presenting for connections.  This is useful for saving and
// later using with `SetTLSCert` to validate future connections.
func (c *Client) FetchTLSCert() (*x509.Certificate, error) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}
	conn, err := tls.Dial("tcp", c.gatewayAddress+":443", tlsConfig)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	certs := conn.ConnectionState().PeerCertificates

	return certs[0], nil
}

// SetRetry sets the retry interval and timeout used when making HTTP requests.
// Setting timeout to zero (or negative) will disable retries (default).
//
// (Note: The client will only attempt retries on network errors (connection
// timed out, etc), not other issues)
func (c *Client) SetRetry(interval time.Duration, timeout time.Duration) {
	c.retryInterval = interval
	c.retryTimeout = timeout
	c.logf("Configured retry settings: interval=%s timeout=%s", interval, timeout)
}

func (c *Client) httpDo(req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error

	start_time := time.Now()
	for {
		attempt_time := time.Now()
		resp, err = c.httpClient.Do(req)
		if _, ok := err.(net.Error); !ok {
			// We only retry on net.Error
			break
		}
		if time.Now().Sub(start_time) >= c.retryTimeout {
			// We've retried as long as we can.  Give up.
			break
		}
		c.logf("Network error fetching API. Retrying... (err=%s)", err)
		// Depending on how long it took to return the error, we may
		// have already used some or all of the time of the retry
		// interval, so figure out how much (if any) is remaining to
		// wait.
		time.Sleep(c.retryInterval - (time.Now().Sub(attempt_time)))
	}
	return resp, err
}

func (c *Client) doHttpRequest(api string, method string, payload []byte, contentType string) ([]byte, error) {
	type errorResponse struct {
		Code    int    `json:"code"`
		Error   string `json:"error"`
		Message string `json:"message"`
	}

	var req *http.Request
	var resp *http.Response
	var err error

	url := url.URL{
		Scheme: "https",
		Host:   c.gatewayAddress,
		Path:   "api/" + api,
	}

	c.logf("Calling API: method=%s url=%s body=%s", method, url.String(), logBody(api, payload, contentType))

	if payload == nil {
		req, err = http.NewRequest(method, url.String(), nil)
		if err != nil {
			return nil, err
		}
	} else {
		req, err = http.NewRequest(method, url.String(), bytes.NewBuffer(payload))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", contentType)
	}

	if strings.HasPrefix(api, "login/") {
		// If we're doing a login API call, don't set the auth cookie,
		// or attempt to retry on auth issues.
		resp, err = c.httpDo(req)
		if err != nil {
			return nil, err
		}
	} else {
		authToken := c.GetAuthToken()
		if authToken != "" {
			cookie := &http.Cookie{
				Name:  "AuthCookie",
				Value: authToken,
			}
			req.AddCookie(cookie)
		}

		resp, err = c.httpDo(req)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode == 401 || resp.StatusCode == 403 {
			// Either we haven't logged in yet (and this API requires login
			// first), or our auth token has expired.  Either way, try
			// logging in (again) and then retry the call.
			resp.Body.Close()
			c.logf("API request returned status %d.  Attempting re-auth...", resp.StatusCode)
			err = c.DoLogin()
			if err != nil {
				return nil, err
			}
			c.logf("Re-auth completed.  Retrying original request.")
			cookie := &http.Cookie{
				Name:  "AuthCookie",
				Value: c.GetAuthToken(),
			}
			req.Header.Del("Cookie")
			req.AddCookie(cookie)
			resp, err = c.httpDo(req)
			if err != nil {
				return nil, err
			}
		}
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		c.logf("Request failed: status=%d body=%s", resp.StatusCode, logBody("", body, resp.Header.Get("Content-Type")))
		errInfo := errorResponse{}
		_ = json.Unmarshal(body, &errInfo)
		return nil, AuthFailure{
			URL:       url,
			ErrorText: errInfo.Error,
			Message:   errInfo.Message,
		}
	} else if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// We got an unexpected response
		c.logf("Request failed: status=%d body=%s", resp.StatusCode, logBody("", body, resp.Header.Get("Content-Type")))
		return body, ApiError{
			URL:        url,
			StatusCode: resp.StatusCode,
			Body:       body,
		}
	}

	c.logf("Request succeeded: status=%d body=%s", resp.StatusCode, logBody(api, body, resp.Header.Get("Content-Type")))

	return body, nil
}

func logBody(api string, body []byte, contentType string) string {
	if strings.HasPrefix(api, "login/") {
		return "(sensitive)"
	} else if body == nil {
		return "(empty)"
	} else if contentType == "application/json" || strings.HasPrefix(contentType, "text/") {
		return string(body)
	} else {
		return fmt.Sprintf("(%d bytes)", len(body))
	}
}

func (c *Client) apiGetJson(api string, result interface{}) error {
	respData, err := c.doHttpRequest(api, http.MethodGet, nil, "")
	if err != nil {
		return err
	}
	err = json.Unmarshal(respData, result)
	if err != nil {
		c.jsonError(api, respData, err)
		return err
	}
	return nil
}

func (c *Client) apiPostJson(api string, payload interface{}, result interface{}) error {
	payloadData, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	respData, err := c.doHttpRequest(api, http.MethodPost, payloadData, "application/json")
	if err != nil {
		return err
	}
	err = json.Unmarshal(respData, result)
	if err != nil {
		c.jsonError(api, respData, err)
		return err
	}
	return nil
}
