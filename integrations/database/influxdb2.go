package database

import (
	"fmt"

	"github.com/jpxor/go-weather-reporter/integrations/weather"
)

type Influxdb2Reporter struct {
}

func (r *Influxdb2Reporter) Report(batch []weather.Weather) error {
	return fmt.Errorf("not implemented")
}
