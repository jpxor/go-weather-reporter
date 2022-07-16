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

package influxdb

import (
	"context"
	"fmt"
	"log"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
	"github.com/jpxor/go-weather-reporter/integrations"
)

var Name = "influxdb2"

type Influxdb2Reporter struct {
	client      influxdb2.Client
	measurement string
	bucket      string
	org         string
	tags        map[string]string
	fields      []string
	logr        *log.Logger
	savedPoints []*write.Point
}

func (r *Influxdb2Reporter) Init(fields []string, config map[string]interface{}) error {
	r.logr = log.New(log.Writer(), "influxdb2 destination: ", log.LstdFlags|log.Lmsgprefix)

	host, ok := config["host"].(string)
	if !ok {
		r.logr.Println("missing required 'host' address")
		return fmt.Errorf("configuration missing required field")
	}
	token, ok := config["token"].(string)
	if !ok {
		r.logr.Println("missing required 'token'")
		return fmt.Errorf("configuration missing required field")
	}
	org, ok := config["org"].(string)
	if !ok {
		r.logr.Println("missing required 'org'")
		return fmt.Errorf("configuration missing required field")
	}
	bucket, ok := config["bucket"].(string)
	if !ok {
		r.logr.Println("missing required 'bucket'")
		return fmt.Errorf("configuration missing required field")
	}
	measurement, ok := config["measurement"].(string)
	if !ok {
		r.logr.Println("missing required 'measurement'")
		return fmt.Errorf("configuration missing required field")
	}
	tags, ok := config["tags"].(map[string]string)
	if !ok {
		r.logr.Println("missing 'tags', leaving blank")
		tags = make(map[string]string)
	}

	r.client = influxdb2.NewClient(host, token)
	r.org = org
	r.bucket = bucket
	r.measurement = measurement
	r.tags = tags
	r.fields = fields

	r.logr.Println("Initialized!")
	return nil
}

func (r *Influxdb2Reporter) Close() {
	r.client.Close()
}

func (r *Influxdb2Reporter) Report(data integrations.Data) error {
	writer := r.client.WriteAPIBlocking(r.org, r.bucket)
	point := influxdb2.NewPoint(
		r.measurement, r.tags, filterDataFields(
			r.fields,
			data.Fields,
		), data.Time,
	)

	// save points so that they can be resubmitted in case
	// of error (ie temporary lost connection)
	r.savedPoints = append(r.savedPoints, point)

	err := writer.WritePoint(context.Background(), r.savedPoints...)
	if err != nil {
		r.logr.Println("influxdb2 failed to WritePoint", err)
		return err
	}

	// erase points that were successfully written
	r.savedPoints = []*write.Point{}
	return err
}

func filterDataFields(fkeys []string, dataFields map[string]integrations.Field) map[string]interface{} {
	fields := make(map[string]interface{})
	for _, field := range fkeys {
		fields[field] = dataFields[field].Value
	}
	return fields
}
