//     go-weather-reporter: pull from weather service, push to database
//     Met Norway integration
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

// Met Norway how-to use api: https://api.met.no/doc/locationforecast/HowTO
// Met Norway data license: https://api.met.no/doc/License

package metno

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	. "github.com/jpxor/go-weather-reporter/integrations/weather"
	. "github.com/jpxor/go-weather-reporter/pkg/httphelper"
)

type CachedResult struct {
	Result       *MetNoResponse
	Expires      time.Time
	LastModified time.Time
}

type MetNoService struct {
	client          *http.Client
	logr            *log.Logger
	cache           map[string]CachedResult
	previousRequest time.Time
}

func NewMetnoWeatherService(logr *log.Logger) *MetNoService {
	client := SimpleClient(10 * time.Second)
	return &MetNoService{
		client:          client,
		logr:            logr,
		cache:           make(map[string]CachedResult),
		previousRequest: time.Unix(0, 0),
	}
}

func (w *MetNoService) Query(loc Location) (*Weather, error) {
	w.logr.Println("querying MET Norway")

	forcast, err := w.locationForecast(w.client, loc.Latitude, loc.Longitude, int(loc.Altitude))
	if err != nil {
		w.logr.Println("metno.LocationForcast failed")
		return nil, err
	}
	instants := forcast.Properties.Timeseries[0].Data.Instant.Details
	data := forcast.Properties.Timeseries[0].Data

	ret := Weather{
		Time: forcast.Properties.Timeseries[0].Time,
		Location: Location{
			Latitude:  forcast.Geometry.Coordinates[0],
			Longitude: forcast.Geometry.Coordinates[1],
			Altitude:  forcast.Geometry.Coordinates[2],
		},
		Measurements: []Measurement{
			{
				Name:  Temperature,
				Value: instants.AirTemperature,
				Unit:  Celcius,
			}, {
				Name:  RelHumidity,
				Value: instants.RelHumidity,
				Unit:  Percent,
			}, {
				Name:  Pressure,
				Value: instants.AirPressure,
				Unit:  HectoPascal,
			}, {
				Name:  Precipitation,
				Value: data.Next1Hours.Details.Precipitation,
				Unit:  Millimeters,
			}, {
				Name:  WindSpeed,
				Value: instants.WindSpeed,
				Unit:  MetersPerSecond,
			}, {
				Name:  CloudCover,
				Value: instants.CloudArea,
				Unit:  Percent,
			}, {
				Name:  "wind_direction",
				Value: instants.WindFromDirection,
				Unit:  Degrees,
			},
		},
	}
	return &ret, nil
}

func (w *MetNoService) locationForecast(client *http.Client, lat, lon float64, alt int) (*MetNoResponse, error) {

	// MetNo service expects client side caching
	cacheHit, lastModified := w.cachedResult(lat, lon, alt)
	if cacheHit != nil {
		w.logr.Println("info: metno using cached result (not yet expired)")
		return cacheHit, nil
	}

	// MetNo considers 20 requests/second to be heavy load,
	// we self-throttle by enforcing at least 50ms interval
	// between requests
	w.selfThrottle()

	url := "https://api.met.no/weatherapi/locationforecast/2.0/compact"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		w.logr.Println("error: failed to create http request", err)
		return nil, err
	}

	q := req.URL.Query()
	q.Add("lat", fmt.Sprintf("%.4f", lat))
	q.Add("lon", fmt.Sprintf("%.4f", lon))
	q.Add("altitude", fmt.Sprintf("%d", alt))
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
			w.logr.Println("warning: throttling", url)
			return nil, ClientErrorRetry
		}
		if res.StatusCode == 403 {
			w.logr.Println("error: access forbidden", url)
			w.logr.Println("  |>> possible black-listed or missing User-Agent identifier")
			return nil, ClientErrorFatal
		}
		if res.StatusCode == 400 {
			w.logr.Println("error: Bad Request", req.URL.String())

			buf, err := ioutil.ReadAll(res.Body)
			if err != nil {
				w.logr.Println("error: failed to read response body")
				return nil, err
			}
			w.logr.Println(string(buf))
			return nil, ClientErrorFatal
		}
		w.logr.Println("error: http client error", res.StatusCode, url)
		return nil, ClientErrorRetry
	}

	if RedirectStatus(res.StatusCode) {
		if res.StatusCode != 304 {
			w.logr.Println("warning: unhandled http status", res.StatusCode, url)
		}
		// data not modified since last response,
		// means cache is still valid even if expired
		cachehit, _ := w.cachedResult(lat, lon, alt)
		if cachehit != nil {
			w.logr.Println("info: metno using cached result (data not modified)")

			// check for new expires
			expiresHdr := res.Header.Get("Expires")
			w.logr.Println("value not changed | new expires header:", expiresHdr)

			// reuse existing last-modified time
			lastModHdr := lastModified.Format(time.RFC1123)

			w.setCachedResult(lat, lon, alt, expiresHdr, lastModHdr, cachehit)
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

		result := MetNoResponse{}
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
		w.setCachedResult(lat, lon, alt, expiresHdr, lastModHdr, &result)

		return &result, nil
	}

	return nil, fmt.Errorf("metno unknown error")
}

func cacheKey(lat, lon float64, alt int) string {
	// lat and long uses a single decimal point of precision
	// alt is meters above sea level and uses 20m resolution
	return fmt.Sprintf("%d-%d-%d", int(lat*10), int(lon*10), alt/20)
}

func (w *MetNoService) cachedResult(lat, lon float64, alt int) (*MetNoResponse, time.Time) {
	cachehit := w.cache[cacheKey(lat, lon, alt)]
	if cachehit.Expires.After(time.Now()) {
		return cachehit.Result, cachehit.LastModified
	}
	return nil, cachehit.LastModified
}

func (w *MetNoService) setCachedResult(lat, lon float64, alt int, expiresHdr, lastModHdr string, result *MetNoResponse) {
	var cacheEntry CachedResult
	expires, err := time.Parse(time.RFC1123, expiresHdr)
	if err != nil {
		w.logr.Println("warning: failed to parse http header date")
	} else {
		cacheEntry = CachedResult{
			Expires: expires,
			Result:  result,
		}
	}
	if cacheEntry.Result == nil {
		// set our own expires if the above failed
		cacheEntry = CachedResult{
			Expires: time.Now().Add(10 * time.Minute),
			Result:  result,
		}
	}
	w.cache[cacheKey(lat, lon, alt)] = cacheEntry
}

func (w *MetNoService) selfThrottle() {
	sinceLastReq := time.Now().Sub(w.previousRequest)
	if sinceLastReq < 50*time.Millisecond {
		time.Sleep(50*time.Millisecond - sinceLastReq)
	}
}

type MetNoResponse struct {
	Type     string `json:"type"`
	Geometry struct {
		Type        string    `json:"type"`
		Coordinates []float64 `json:"coordinates"`
	} `json:"geometry"`
	Properties struct {
		Meta struct {
			UpdatedAt time.Time `json:"updated_at"`
			Units     struct {
				AirPressure       string `json:"air_pressure_at_sea_level"`
				AirTemperature    string `json:"air_temperature"`
				CloudArea         string `json:"cloud_area_fraction"`
				Precipitation     string `json:"precipitation_amount"`
				RelHumidity       string `json:"relative_humidity"`
				WindFromDirection string `json:"wind_from_direction"`
				WindSpeed         string `json:"wind_speed"`
			} `json:"units"`
		} `json:"meta"`
		Timeseries []struct {
			Time time.Time `json:"time"`
			Data struct {
				Instant struct {
					Details struct {
						AirPressure       float32 `json:"air_pressure_at_sea_level"`
						AirTemperature    float32 `json:"air_temperature"`
						CloudArea         float32 `json:"cloud_area_fraction"`
						RelHumidity       float32 `json:"relative_humidity"`
						WindFromDirection float32 `json:"wind_from_direction"`
						WindSpeed         float32 `json:"wind_speed"`
					} `json:"details"`
				} `json:"instant"`
				Next1Hours struct {
					Summary struct {
						SymbolCode string `json:"symbol_code"`
					} `json:"summary"`
					Details struct {
						Precipitation float32 `json:"precipitation_amount"`
					} `json:"details"`
				} `json:"next_1_hours"`
				Next6Hours struct {
					Summary struct {
						SymbolCode string `json:"symbol_code"`
					} `json:"summary"`
					Details struct {
						Precipitation float32 `json:"precipitation_amount"`
					} `json:"details"`
				} `json:"next_6_hours"`
				Next12Hours struct {
					Summary struct {
						SymbolCode string `json:"symbol_code"`
					} `json:"summary"`
					Details struct {
						Precipitation float32 `json:"precipitation_amount"`
					} `json:"details"`
				} `json:"next_12_hours"`
			} `json:"data"`
		} `json:"timeseries"`
	} `json:"properties"`
}
