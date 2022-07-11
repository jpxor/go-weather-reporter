package httphelper

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"
)

// HTTP Client helpers
var ServerErrorFatal = fmt.Errorf("server error (FATAL)")
var ServerErrorRetry = fmt.Errorf("server error (EAGAIN)")

var ClientErrorFatal = fmt.Errorf("client error (FATAL)")
var ClientErrorRetry = fmt.Errorf("client error (EAGAIN)")

func ServerErrorStatus(code int) bool {
	return code >= 500
}

func ClientErrorStatus(code int) bool {
	return code >= 400 && code < 500
}

func RedirectStatus(code int) bool {
	return code >= 300 && code < 400
}

func SuccessStatus(code int) bool {
	return code >= 200 && code < 300
}

// from: https://github.com/jackdoe/go-metno/blob/master/metno.go
func SimpleClient(timeout time.Duration) *http.Client {
	var netTransport = &http.Transport{
		Dial: (&net.Dialer{
			Timeout: timeout,
		}).Dial,
		DisableCompression:  false,
		TLSHandshakeTimeout: timeout,
	}
	proxyString := os.Getenv("HTTPS_PROXY")
	if proxyString != "" {
		proxyUrl, err := url.Parse(proxyString)
		if err == nil {
			netTransport.Proxy = http.ProxyURL(proxyUrl)
		}
	}
	var netClient = &http.Client{
		Timeout:   timeout,
		Transport: netTransport,
	}
	return netClient
}
