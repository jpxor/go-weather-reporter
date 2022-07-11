package weather

import "time"

type Interface interface {
	Query(location Location) (*Weather, error)
}

type Weather struct {
	Time         time.Time     `json:"time"`
	Location     Location      `json:"location"`
	Measurements []Measurement `json:"measurements"`
}

type Measurement struct {
	Name  string  `json:"name"`
	Value float32 `json:"value"`
	Unit  string  `json:"unit"`
}

type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Altitude  float64 `json:"altitude"`
}

const (
	Pressure      = "pressure"
	Temperature   = "temperature"
	RelHumidity   = "relative_humidity"
	CloudCover    = "cloud_cover"
	Precipitation = "precipitation"
	WindSpeed     = "wind_speed"
)

const (
	Celcius     = "celsius"
	Farenheight = "fahrenheit"
	Kelvin      = "kelvin"

	Percent = "%"
	Degrees = "degrees"
	Radians = "radians"

	Millimeters = "mm"
	Centimeters = "cm"
	Inches      = "in"

	MetersPerSecond   = "m/s"
	KilometersPerHour = "Kph"
	MilesPerHour      = "mph"

	HectoPascal = "hPa"
	Bars        = "bar"
	Atmospheres = "atm"

	Text = "text"
)
