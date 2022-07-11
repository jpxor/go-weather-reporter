package internal

import (
	"log"
	"os"
	"time"

	"github.com/jpxor/go-weather-reporter/integrations/weather"
)

func Run(config Config, logr *log.Logger) {

	if config.Start {
		logr.Println("started")
		config.WeatherService = weather.NewMetnoWeatherService(logr)

		for {
			result, err := config.WeatherService.Query(weather.Location{
				Latitude:  45.357437,
				Longitude: -75.805931,
				Altitude:  70,
			})
			if err != nil {
				if err == weather.ClientErrorFatal {
					os.Exit(1)
				}
			} else {
				logr.Println(result)
			}
			time.Sleep(time.Second)
		}

	}
}
