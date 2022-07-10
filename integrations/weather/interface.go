package weather

import "time"

type Interface interface {
	Query(window Window, location Location) (*Weather, error)
}

// strongly based on:
// https://github.com/jackdoe/go-metno/blob/master/metno.go
// which was generated from json output of api.met.no
type Weather struct {
	Created       time.Time     `json:"created"`
	Window        Window        `json:"window"`
	Location      Location      `json:"location"`
	Fog           Fog           `json:"fog"`
	Cloudiness    Cloudiness    `json:"cloudiness"`
	Pressure      Pressure      `json:"pressure"`
	WindDirection WindDirection `json:"windDirection"`
	WindGust      WindGust      `json:"windGust"`
	WindSpeed     WindSpeed     `json:"windSpeed"`
	Temperature   Temperature   `json:"temperature"`
	Dewpoint      Dewpoint      `json:"dewpoint"`
}

type Window struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}

type Location struct {
	Longitude float64 `json:"longitude,string"`
	Altitude  float64 `json:"altitude,string"`
	Latitude  float64 `json:"latitude,string"`
}

type Fog struct {
	ID      string  `json:"id"`
	Percent float64 `json:"percent,string"`
}

type Pressure struct {
	ID    string  `json:"id"`
	Unit  string  `json:"unit"`
	Value float64 `json:"value,string"`
}

type Cloudiness struct {
	Percent float64 `json:"percent,string"`
	ID      string  `json:"id"`
}

type WindDirection struct {
	Deg  float64 `json:"deg,string"`
	Name string  `json:"name"`
	ID   string  `json:"id"`
}

type WindGust struct {
	Mps float64 `json:"mps,string"`
	ID  string  `json:"id"`
}

type WindSpeed struct {
	Beaufort string  `json:"beaufort"`
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Mps      float64 `json:"mps,string"`
}

type Temperature struct {
	Value float64 `json:"value,string"`
	Unit  string  `json:"unit"`
	ID    string  `json:"id"`
}

type Dewpoint struct {
	ID    string  `json:"id"`
	Value float64 `json:"value,string"`
	Unit  string  `json:"unit"`
}

type Humidity struct {
	Value float64 `json:"value,string"`
	Unit  string  `json:"unit"`
}
