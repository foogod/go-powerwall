package powerwall

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var logFunc = func(v ...interface{}) {}

func SetLogFunc(f func(...interface{})) {
	logFunc = f
}

var errFunc = func(string, error) {}

func SetErrFunc(f func(string, error)) {
	errFunc = f
}

type Client struct {
	gatewayAddress       string
	gatewayLoginEmail    string
	gatewayLoginPassword string
	httpClient           http.Client
	token_ch             chan string
	auth_ch              chan *authMessage
}

func (c *Client) logf(format string, v ...interface{}) {
	logFunc(fmt.Sprintf("{%p} ", c) + fmt.Sprintf(format, v...))
}

func (c *Client) jsonError(api string, data []byte, err error) {
	msg := fmt.Sprintf("Error unmarshalling '%s' response %s", api, string(data))
	errFunc(msg, err)
}

func NewClient(gatewayAddress string, gatewayLoginEmail string, gatewayLoginPassword string) *Client {
	tr := &http.Transport{
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
	httpClient := http.Client{
		Transport: tr,
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

	go c.authManager()

	c.logf("New powerwall client created: gateway_address=%s email=%s", gatewayAddress, gatewayLoginEmail)
	return c
}

func (c *Client) SetTLSCert(cert *x509.Certificate) {
	certPool := x509.NewCertPool()
	certPool.AddCert(cert)

	c.httpClient.Transport.(*http.Transport).TLSClientConfig.InsecureSkipVerify = false
	c.httpClient.Transport.(*http.Transport).TLSClientConfig.RootCAs = certPool
	c.logf("Set TLS cert for validation: %s", cert.Subject)
}

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
		resp, err = c.httpClient.Do(req)
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

		resp, err = c.httpClient.Do(req)
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
			resp, err = c.httpClient.Do(req)
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
