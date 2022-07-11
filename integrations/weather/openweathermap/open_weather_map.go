//     go-weather-reporter: pull from weather service, push to database
//     OpenWeatherMap integration
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

// https://openweathermap.org/
// api reference: https://openweathermap.org/current

package openweathermap

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	. "github.com/jpxor/go-weather-reporter/integrations/weather"
	. "github.com/jpxor/go-weather-reporter/pkg/httphelper"
)

type CachedResult struct {
	Result       *OpenWeatherResponse
	Expires      time.Time
	LastModified time.Time
}

type OpenWeatherService struct {
	client          *http.Client
	logr            *log.Logger
	cache           map[string]CachedResult
	previousRequest time.Time
	ApiKey          string
	Language        string
}

func NewWeatherService(apiKey, lang string, logr *log.Logger) *OpenWeatherService {
	client := SimpleClient(10 * time.Second)
	return &OpenWeatherService{
		client:          client,
		logr:            logr,
		cache:           make(map[string]CachedResult),
		previousRequest: time.Unix(0, 0),
		ApiKey:          apiKey,
		Language:        lang,
	}
}

func (w *OpenWeatherService) Query(loc Location) (*Weather, error) {
	w.logr.Println("querying OpenWeather")

	current, err := w.currentWeatherQuery(w.client, loc.Latitude, loc.Longitude)
	if err != nil {
		w.logr.Println("openweather.currentWeatherQuery failed")
		return nil, err
	}

	ret := Weather{
		Time: time.Unix(int64(current.Time), 0),
		Location: Location{
			Latitude:  current.Location.Latitude,
			Longitude: current.Location.Longitude,
		},
		Measurements: []Measurement{
			{
				Name:  Temperature,
				Value: current.Main.Temperature,
				Unit:  Celcius,
			}, {
				Name:  "FeelsLike",
				Value: current.Main.FeelsLike,
				Unit:  Celcius,
			}, {
				Name:  RelHumidity,
				Value: current.Main.RelHumidity,
				Unit:  Percent,
			}, {
				Name:  Pressure,
				Value: current.Main.Pressure,
				Unit:  HectoPascal,
			}, {
				Name:  WindSpeed,
				Value: current.Wind.Speed,
				Unit:  MetersPerSecond,
			}, {
				Name:  CloudCover,
				Value: current.Clouds.All,
				Unit:  Percent,
			},
		},
	}

	return &ret, nil
}

func (w *OpenWeatherService) currentWeatherQuery(client *http.Client, lat, lon float64) (*OpenWeatherResponse, error) {

	cacheHit, lastModified := w.cachedResult(lat, lon)
	if cacheHit != nil {
		w.logr.Println("info: openwweather using cached result (not yet expired)")
		return cacheHit, nil
	}
	w.selfThrottle()

	// NOTE: The endpoint for paid subscription plans is different
	url := "https://api.openweathermap.org/data/2.5/weather"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		w.logr.Println("error: failed to create http request", err)
		return nil, err
	}

	q := req.URL.Query()
	q.Add("lat", fmt.Sprintf("%.4f", lat))
	q.Add("lon", fmt.Sprintf("%.4f", lon))
	q.Add("units", "metric")
	q.Add("lang", "en")
	q.Add("appid", w.ApiKey)
	req.URL.RawQuery = q.Encode()

	req.Header.Add("Accept", "application/json")
	req.Header.Add("If-Modified-Since", lastModified.Format(time.RFC1123))
	req.Header.Set("User-Agent", "go-weather-reporter client (https://github.com/jpxor/go-weather-reporter)")

	res, err := client.Do(req)
	if err != nil {
		w.logr.Println("error: failed to send http request", err)
		return nil, err
	}
	defer res.Body.Close()

	// request sent, so track the time for self throttling
	w.previousRequest = time.Now()

	if ServerErrorStatus(res.StatusCode) {
		w.logr.Println("warning: server error", url)
		return nil, ServerErrorRetry
	}

	if ClientErrorStatus(res.StatusCode) {
		if res.StatusCode == 429 {
			w.logr.Println("warning: your OpenWeatherMap.org account is temporary blocked due to exceeding of requests limitation of your subscription type", url)
			return nil, ClientErrorFatal
		}
		if res.StatusCode == 403 {
			w.logr.Println("error: access forbidden", url)
			return nil, ClientErrorFatal
		}
		if res.StatusCode == 400 {
			w.logr.Println("error: Bad Request", req.URL.String())
			printResponseBody(w.logr, res.Body)
			return nil, ClientErrorFatal
		}
		w.logr.Println("error: http client error", res.StatusCode, url)
		printResponseBody(w.logr, res.Body)

		w.logr.Println("Note: if you recently created the OpenWeatherMap api-key, try again in a few minutes")
		return nil, ClientErrorRetry
	}

	if RedirectStatus(res.StatusCode) {
		if res.StatusCode != 304 {
			w.logr.Println("warning: unhandled http status", res.StatusCode, url)
		}
		// data not modified since last response,
		// means cache is still valid even if expired
		cachehit, _ := w.cachedResult(lat, lon)
		if cachehit != nil {
			w.logr.Println("info: openweather using cached result (data not modified)")

			// check for new expires
			expiresHdr := res.Header.Get("Expires")
			w.logr.Println("value not changed | new expires header:", expiresHdr)

			// reuse existing last-modified time
			lastModHdr := lastModified.Format(time.RFC1123)

			w.setCachedResult(lat, lon, expiresHdr, lastModHdr, cachehit)
			return cachehit, nil
		}
	}

	if SuccessStatus(res.StatusCode) {
		if res.StatusCode == 203 {
			w.logr.Println("warning: depreciated service or api:", url)
			w.logr.Println("  |>> options: update, create pull request, or open an issue")
			w.logr.Println("  |>> see: https://github.com/jpxor/go-weather-reporter/issues")
		}

		buf, err := ioutil.ReadAll(res.Body)
		if err != nil {
			w.logr.Println("error: failed to read response from", url)
			return nil, err
		}

		result := OpenWeatherResponse{}
		err = json.Unmarshal(buf, &result)
		if err != nil {
			w.logr.Println("error: failed to parse response from", url)
			w.logr.Println(err)
			return nil, err
		}

		// cache result
		expiresHdr := res.Header.Get("Expires")
		lastModHdr := res.Header.Get("Last-Modified")
		w.logr.Println("caching result | expires header:", expiresHdr)
		w.setCachedResult(lat, lon, expiresHdr, lastModHdr, &result)

		return &result, nil
	}

	return nil, fmt.Errorf("openweather unknown error")
}

func printResponseBody(logr *log.Logger, body io.ReadCloser) {
	buf, err := ioutil.ReadAll(body)
	if err != nil {
		logr.Println("error: failed to read response body")
	}
	logr.Println(string(buf))
}

func cacheKey(lat, lon float64) string {
	// lat and long uses a single decimal point of precision
	// alt is meters above sea level and uses 20m resolution
	return fmt.Sprintf("%d-%d", int(lat*10), int(lon*10))
}

func (w *OpenWeatherService) cachedResult(lat, lon float64) (*OpenWeatherResponse, time.Time) {
	cachehit := w.cache[cacheKey(lat, lon)]
	if cachehit.Expires.After(time.Now()) {
		return cachehit.Result, cachehit.LastModified
	}
	return nil, cachehit.LastModified
}

func (w *OpenWeatherService) setCachedResult(lat, lon float64, expiresHdr, lastModHdr string, result *OpenWeatherResponse) {
	var cacheEntry CachedResult
	expires, err := time.Parse(time.RFC1123, expiresHdr)
	if err == nil {
		cacheEntry = CachedResult{
			Expires: expires,
			Result:  result,
		}
	}
	if cacheEntry.Result == nil {
		// set our own expires if the above failed
		// From https://openweathermap.org
		//    " First, we recommend making API calls no more than once in 10 minutes for
		//      each location, whether you call it by city name, geographical coordinates
		//      or by zip code. The update frequency of the OpenWeather model is not
		//      higher than once in 10 minutes. "
		cacheEntry = CachedResult{
			Expires: time.Now().Add(10 * time.Minute),
			Result:  result,
		}
	}
	w.cache[cacheKey(lat, lon)] = cacheEntry
}

func (w *OpenWeatherService) selfThrottle() {
	// openweathermap.org free-tier allows 60 calls per minute,
	// so we self-throttle by enforcing at least 1 second interval
	// between requests
	sinceLastReq := time.Now().Sub(w.previousRequest)
	if sinceLastReq < 1*time.Second {
		time.Sleep(1*time.Second - sinceLastReq)
	}
}

type OpenWeatherResponse struct {
	Location struct {
		Longitude float64 `json:"lon"`
		Latitude  float64 `json:"lat"`
	} `json:"coord"`
	Weather []struct {
		ID          int    `json:"id"`
		Main        string `json:"main"`
		Description string `json:"description"`
		IconID      string `json:"icon"`
	} `json:"weather"`
	Base string `json:"base"`
	Main struct {
		Temperature    float32 `json:"temp"`
		FeelsLike      float32 `json:"feels_like"`
		MinTemperature float32 `json:"temp_min"`
		MaxTemperature float32 `json:"temp_max"`
		Pressure       float32 `json:"pressure"`
		RelHumidity    float32 `json:"humidity"`
	} `json:"main"`
	Visibility float32 `json:"visibility"`
	Wind       struct {
		Speed     float32 `json:"speed"`
		Direction float32 `json:"deg"`
		Gust      float32 `json:"gust"`
	} `json:"wind"`
	Clouds struct {
		All float32 `json:"all"`
	} `json:"clouds"`
	Time int `json:"dt"`
	Sys  struct {
		Type    int    `json:"type"`
		ID      int    `json:"id"`
		Country string `json:"country"`
		Sunrise int    `json:"sunrise"`
		Sunset  int    `json:"sunset"`
	} `json:"sys"`
	Timezone int    `json:"timezone"`
	ID       int    `json:"id"`
	Name     string `json:"name"`
	COD      int    `json:"cod"`
}
