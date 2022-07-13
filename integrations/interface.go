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

package integrations

import "time"

type Field struct {
	Value interface{}
	Unit  string
}

type Data struct {
	Time   time.Time
	Fields map[string]Field
}

type SourceInterface interface {
	Init(config map[string]interface{}) error
	Query() (Data, error)
}

type DestinationInterface interface {
	Init(configInfo map[string]interface{}) error
	Report(Data) error
}
