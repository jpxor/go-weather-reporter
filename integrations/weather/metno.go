package weather

import "fmt"

type MetNoWeather struct {
}

func (w *MetNoWeather) Query(window Window, location Location) (*Weather, error) {
	return nil, fmt.Errorf("not implemented")
}
