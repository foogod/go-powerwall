// Functions for authentication and login:
//
// (Note: DoLogin generally does not need to be called explicitly)
//
//   (*Client) DoLogin()
//   (*Client) GetAuthToken()
//   (*Client) SetAuthToken(token string)
//
package powerwall

import (
	"errors"
)

const (
	cmd_DO_LOGIN int = iota
	cmd_CHECK_LOGIN
	cmd_SET_TOKEN
)

type authMessage struct {
	action    int
	email     string
	password  string
	token     string
	result_ch chan error
}

// DoLogin logs into the Powerwall gateway and obtains an auth token which can
// be used for subsequent API calls.  Note that you should not normally need to
// call this explicitly.  The library will automatically attempt to login
// anyway if a call is made which requires authentication and it is not already
// successfully logged in.
func (c *Client) DoLogin() error {
	action := authMessage{
		action:    cmd_DO_LOGIN,
		email:     c.gatewayLoginEmail,
		password:  c.gatewayLoginPassword,
		result_ch: make(chan error),
	}
	c.auth_ch <- &action
	return <-action.result_ch
}

func (c *Client) checkLogin() error {
	action := authMessage{
		action:    cmd_CHECK_LOGIN,
		email:     c.gatewayLoginEmail,
		password:  c.gatewayLoginPassword,
		result_ch: make(chan error),
	}
	c.auth_ch <- &action
	return <-action.result_ch
}

// GetAuthToken returns the current auth token in use.  This can be saved and
// then passed to SetAuthToken on later connections to re-use the same token
// across Clients.
func (c *Client) GetAuthToken() string {
	return <-c.token_ch
}

// SetAuthToken sets the provided string as the new auth token to use for
// subsequent API calls.
func (c *Client) SetAuthToken(token string) {
	c.auth_ch <- &authMessage{action: cmd_SET_TOKEN, token: token}
	// Wait until we are sure the manager is returning the updated token before returning.
	for {
		t := <-c.token_ch
		if t == token {
			break
		}
	}
}

func (c *Client) authManager() {
	var authToken = ""

	for {
		// We do a double-select here because we want to ensure that
		// auth messages are always processed with a higher priority
		// than token requests, so we always check that queue first,
		// and only if there's no messages waiting, then wait on both
		// queues as normal.
		select {
		case msg := <-c.auth_ch:
			c.doAuthMsg(msg, &authToken)
			continue
		default:
		}
		select {
		case msg := <-c.auth_ch:
			c.doAuthMsg(msg, &authToken)
			continue
		case c.token_ch <- authToken:
		}
	}
}

func (c *Client) doAuthMsg(msg *authMessage, authToken *string) {
	var err error

	switch msg.action {
	case cmd_SET_TOKEN:
		*authToken = msg.token
		c.logf("Set auth token")
	case cmd_DO_LOGIN:
		*authToken, err = c.performLogin(msg.email, msg.password)
		msg.result_ch <- err
	case cmd_CHECK_LOGIN:
		if *authToken == "" {
			*authToken, err = c.performLogin(msg.email, msg.password)
		} else {
			err = nil
		}
		msg.result_ch <- err
	}
}

type loginData struct {
	Username   string `json:"username"`
	Email      string `json:"email"`
	Password   string `json:"password"`
	ForceSmOff bool   `json:"force_sm_off"`
}

type loginResponse struct {
	Email     string   `json:"email"`
	FirstName string   `json:"firstname"`
	LastName  string   `json:"lastname"`
	Roles     []string `json:"roles"`
	Token     string   `json:"token"`
	Provider  string   `json:"provider"`
	loginTime string   `json:"loginTime"`
}

func (c *Client) performLogin(email, password string) (string, error) {
	c.logf("Attempting login...")
	ld := loginData{
		Username:   "customer",
		Email:      email,
		Password:   password,
		ForceSmOff: false,
	}
	resp := loginResponse{}
	err := c.apiPostJson("login/Basic", ld, &resp)

	// Check for presence of a Token first, because if there was some issue
	// unmarshalling the full response, it will return an error, but it may
	// have been able to unmarshal enough to extract a valid token anyway,
	// which is really all we need.
	if resp.Token != "" {
		// We got back a token.  We're good!
		c.logf("Login successful")
		return resp.Token, nil
	} else if err != nil {
		c.logf("Login failed: %s", err)
		return "", err
	} else {
		// No error, but also no token?
		c.logf("Login successful but no token returned?")
		return "", errors.New("No auth token returned from login API call")
	}
}
