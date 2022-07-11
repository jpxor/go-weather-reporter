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
