package internal

import (
	"time"

	"github.com/jpxor/go-weather-reporter/integrations/database"
	"github.com/jpxor/go-weather-reporter/integrations/weather"
)

type Config struct {
	Start          bool
	QueryInterval  time.Duration
	ReportInterval time.Duration
	Database       struct {
		HostPath string
		Username string
		Password string
		Token    string
	}
	WeatherService weather.Interface
	DataService    database.Interface
}
