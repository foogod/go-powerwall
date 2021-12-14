package main

import (
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"strings"

	"github.com/jessevdk/go-flags"

	"github.com/foogod/go-powerwall"
)

var options struct {
	Debug     bool   `long:"debug" description:"Enable debug messages"`
	Address   string `long:"address" required:"true" description:"IP address or hostname of Powerwall gateway (required)"`
	Email     string `long:"email" description:"Email address to use when logging in"`
	Password  string `long:"password" description:"Password to use when logging in"`
	AuthCache string `long:"authcache" description:"Filename to store/load auth token"`
	CertFile  string `long:"certfile" description:"Filename of TLS certificate to use for validation"`
	Args      struct {
		Command string   `positional-arg-name:"command" description:"One of 'status', 'login', 'site_info', 'fetchcert', 'aggregates', 'meter', 'system_status', 'grid_faults', 'grid_status', 'soe', 'operation', 'sitemaster', 'networks'"`
		Args    []string `positional-arg-name:"args" description:"Optional arguments depending on command"`
	} `positional-args:"true" required:"true"`
}

func logDebug(v ...interface{}) {
	log.Debug(v...)
}

func logError(msg string, err error) {
	log.WithFields(log.Fields{"err": err}).Error(msg)
}

func main() {
	var err error

	_, err = flags.Parse(&options)
	if err != nil {
		os.Exit(1)
	}

	if options.Debug {
		log.SetLevel(log.DebugLevel)
	}
	powerwall.SetLogFunc(logDebug)
	powerwall.SetErrFunc(logError)

	c := powerwall.NewClient(options.Address, options.Email, options.Password)

	if options.CertFile != "" && options.Args.Command != "fetchcert" {
		pemCert, err := ioutil.ReadFile(options.CertFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Cannot read cert file: %s\n", err)
			os.Exit(2)
		}
		block, _ := pem.Decode(pemCert)
		if block == nil || block.Type != "CERTIFICATE" {
			fmt.Fprintln(os.Stderr, "Unable to decode cert file.  Is it in PEM format?")
			os.Exit(2)
		}
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading cert file: %s\n", err)
			os.Exit(2)
		}

		c.SetTLSCert(cert)
	}

	authToken := ""
	if options.AuthCache != "" {
		authdata, err := ioutil.ReadFile(options.AuthCache)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				// This is ok
			} else {
				fmt.Fprintf(os.Stderr, "Cannot read authcache file: %s\n", err)
				os.Exit(2)
			}
		}
		authToken = strings.TrimSpace(string(authdata))
		c.SetAuthToken(authToken)
	}

	switch options.Args.Command {
	case "fetchcert":
		cert, err := c.FetchTLSCert()
		if err != nil {
			panic(err)
		}
		pemCert := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw})
		err = ioutil.WriteFile(options.CertFile, pemCert, 0644)
		if err != nil {
			panic(err)
		}
	case "status":
		result, err := c.GetStatus()
		if err != nil {
			panic(err)
		}
		writeResult(result)
	case "login":
		err := c.DoLogin()
		if err != nil {
			panic(err)
		}
		// If there's no authcache file, just print the auth token to stdout
		if options.AuthCache == "" {
			fmt.Println(c.GetAuthToken())
		}
	case "site_info":
		result, err := c.GetSiteInfo()
		if err != nil {
			panic(err)
		}
		writeResult(result)
	case "aggregates":
		result, err := c.GetMetersAggregates()
		if err != nil {
			panic(err)
		}
		writeResult(result)
	case "meter":
		result, err := c.GetMeters(options.Args.Args[0])
		if err != nil {
			panic(err)
		}
		writeResult(result)
	case "system_status":
		result, err := c.GetSystemStatus()
		if err != nil {
			panic(err)
		}
		writeResult(result)
	case "grid_faults":
		result, err := c.GetGridFaults()
		if err != nil {
			panic(err)
		}
		writeResult(result)
	case "grid_status":
		result, err := c.GetGridStatus()
		if err != nil {
			panic(err)
		}
		writeResult(result)
	case "soe":
		result, err := c.GetSOE()
		if err != nil {
			panic(err)
		}
		writeResult(result)
	case "operation":
		result, err := c.GetOperation()
		if err != nil {
			panic(err)
		}
		writeResult(result)
	case "sitemaster":
		result, err := c.GetSitemaster()
		if err != nil {
			panic(err)
		}
		writeResult(result)
	case "networks":
		result, err := c.GetNetworks()
		if err != nil {
			panic(err)
		}
		writeResult(result)
	default:
		fmt.Fprintf(os.Stderr, "Error: Unknown command: %v\n", options.Args.Command)
		os.Exit(3)
	}

	newAuthToken := c.GetAuthToken()
	if newAuthToken != authToken && options.AuthCache != "" {
		// Auth token has changed.  Write it out to the cache file.
		err := os.WriteFile(options.AuthCache, []byte(newAuthToken), 0600)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Cannot write to authcache file: %s\n", err)
		}
	}
}

func writeResult(value interface{}) {
	b, err := json.MarshalIndent(value, "", "    ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))
}
