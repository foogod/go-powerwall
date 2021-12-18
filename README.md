# go-powerwall

A Go library for communicating with Tesla Powerwall appliances via the local-network API.

---

***Many thanks to [Vince Loschiavo](https://github.com/vloschiavo) and other contributors to https://github.com/vloschiavo/powerwall2 for providing a lot of the information to make this possible!***

**Note:** The Tesla powerwall interface is an undocumented API which is not supported by Tesla, and could change at any time.  This library has been tested on devices running version 21.39 and 21.44 of the firmware only, at this time.

**This library is incomplete.**  Pull requests to add support for more functions are always welcome!

---

This library is intended to allow clients to easily pull information from and send commands to a Tesla Powerwall via a Tesla Home Energy Gateway 2.  Note that this is the internal local-network API exposed by the device itself.  It is not the same as the [Tesla Owner API](https://tesla-api.timdorr.com/), which allows you to control the device via Tesla's servers over the Internet.  To use this API, your Powerwall Energy Gateway must be connected to a local network via WiFi or Ethernet, and accessible via that network.

## Basic Usage

First, create a new client instance by calling `powerwall.NewClient`, providing the hostname or IP address of the gateway appliance, along with the email address and password which should be used to login.  (If you have not already, it is recommended that you first login to the device using a web browser and change your password, as detailed [on Tesla's website](https://www.tesla.com/support/energy/powerwall/own/monitoring-from-home-network).)  Once you have a client object, you can use its various functions to retrieve information from the device, etc.

```go
import (
	"fmt"
	"github.com/foogod/go-powerwall"
)

func main() {
	client := powerwall.NewClient("192.168.123.45", "teslaguy@example.com", "MySuperSecretPassword!")
	result, err := client.GetStatus()
	if err != nil {
		panic(err)
	}
	fmt.Printf("The gateway's ID number is: %s\nIt is running version: %s\n", result.Din, result.Version)
}
```

The client will automatically login to the device as needed, and will remember and re-use the auth-token between calls.  It will also automatically re-login if necessary (i.e. if the token expires).

## TLS Certificates

The Tesla gateway uses a self-signed certificate, which means that it shows up as invalid by default (because it is not signed by any known authority).  For this reason, the default behavior of the client is to not try to validate the TLS certificate when connecting.  This works, but it is insecure, as it is possible for someone else to impersonate the gateway instead (a "man in the middle attack").  If a more secure configuration is desired, the library does support a way to do full TLS validation, but you will need to provide it with a copy of the certificate to validate against after creating the client, using the `SetTLSCert` function.

```go
	client := powerwall.NewClient("192.168.123.45", "someguy@example.com", "MySuperSecretPassword!")
	client.SetTLSCert(cert)
```

The certificate can be obtained using command-line utilities such as `openssl`; however, the client does also have a handy function for retrieving the certificate from a Powerwall gateway directly, as well:

```go
	cert, err := client.FetchTLSCert()
```

The typical use-case of this would be to provide a special option or command in your program to fetch and store the certificate in a file initially, and from then on read it from the file and use that to set the certificate when creating any new clients going forward, before performing any API calls.  For an example of this, see the `--certfile` and `fetchcert` options of the [powerwall-cmd](cmd/powerwall-cmd/main.go) sample program in this repo.

## Retrying requests

Tesla's Powerwall appliances seem to have some issues staying reliably connected to WiFi networks (for me, at least), and will periodically become disconnected for a few seconds and then reconnect.  This can cause a problem if you happen to hit your API request at the wrong time, as you will just end up with a network error instead.

To work around this, the client can be configured to retry HTTP requests for a given amount of time before actually giving up, if they are failing because of network connection errors.  This can be done by calling the SetRetry function with the desired retry interval and timeout:

```go
	// Retry attempts every second, giving up after a minute of failed attempts.
	interval := time.ParseDuration("1s")
	timeout := time.ParseDuration("60s")
	client.SetRetry(interval, timeout)
```

The client will wait at least the specified interval between attempts (but it may sometimes be longer if, for example, a connection timeout takes longer than that interval just to return the error).  It will keep trying until the timeout has been exceeded (doing however many retries it can fit within that period).

This behavior is disabled by default.  Setting the timeout to zero (or negative) will also disable all retries.  (Note that setting interval to zero (or negative) is also allowed, but will result in the library attempting to retry as fast as possible, which may produce excessive network traffic or CPU usage, so it is not generally advised.)

## Saving and re-using the auth token

If you are making a program which needs to regularly create new clients (such as a command-line utility which gets run on a regular basis to collect stats and then exit, etc), it may be desirable to save the auth token after login so that it can be re-used later.  This can be done using the `GetAuthToken` and `SetAuthToken` functions:

```go
	token := client.GetAuthToken()
	client.SetAuthToken(token)
```

(Note that the client must have already performed a login before you will be able to retrieve a valid token with `GetAuthToken`.  This will happen automatically if you attempt to fetch from an API which requires authentication, or you can use `DoLogin` initially to force a login operation first.)

Keep in mind that the client will automatically re-login (and thus generate a new auth token) if the provided one is invalid (it has expired, etc).  It is therefore a good idea to check periodically whether the token has changed (using GetAuthToken), and if so update your saved copy as well.

For an example of this, see the `--authcache` option of the [powerwall-cmd](cmd/powerwall-cmd/main.go) sample program in this repo.

## Logging

If something is not working properly, it may be useful to get debug logging of what the `powerwall` library is doing behind the scenes (including HTTP requests/responses, etc).  You can register a logging function for this purpose using `powerwall.SetLogFunc`:

```go
func myLogFunc(v ...interface{}) {
        msg := fmt.Sprintf(v...)
	fmt.Printf("powerwall: %s\n", msg)
}

func main() {
	powerwall.SetLogFunc(logDebug)

	(...)
}
```

(The arguments to the log function are the same as for Sprintf/etc.  Note, however, that generated log lines are not terminated with "\n", so you will need to add a newline if you are sending them to something like `fmt.Printf` directly.)

