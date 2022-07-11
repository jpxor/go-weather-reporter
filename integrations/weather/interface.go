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
