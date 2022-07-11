//     go-weather-reporter: pull from weather service, push to database
//     Copyright (C) 2022 Josh Simonot
//
//     This program is free software: you can redistribute it and/or modify
//     it under the terms of the GNU General Public License as published by
//     the Free Software Foundation, either version 3 of the License, or
//     (at your option) any later version.
//
//     This program is distributed in the hope that it will be useful,
//     but WITHOUT ANY WARRANTY; without even the implied warranty of
//     MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//     GNU General Public License for more details.
//
//     You should have received a copy of the GNU General Public License
//     along with this program.  If not, see <https://www.gnu.org/licenses/>.

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

//********************************************************************//
//** The following function is separately licensed                  **//
//** Copied From:                                                   **//
//**     https://github.com/jackdoe/go-metno/blob/master/metno.go   **//
//********************************************************************//
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
