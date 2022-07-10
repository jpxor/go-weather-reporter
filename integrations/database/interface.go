package database

import "github.com/jpxor/go-weather-reporter/integrations/weather"

type Interface interface {
	Report(batch []weather.Weather) error
}
