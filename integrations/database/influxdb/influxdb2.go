//     go-weather-reporter: pull from weather service, push to database
//     Influxdb2 integration
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

package database

import (
	"fmt"

	"github.com/jpxor/go-weather-reporter/integrations/weather"
)

type Influxdb2Reporter struct {
}

func NewReporter() *Influxdb2Reporter {
	return &Influxdb2Reporter{}
}

func (r *Influxdb2Reporter) Report(batch []*weather.Weather) error {
	return fmt.Errorf("not implemented")
}
