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
	"log"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
	"github.com/jpxor/go-weather-reporter/integrations/database"
	"github.com/jpxor/go-weather-reporter/integrations/weather"
)

type Influxdb2Reporter struct {
	client      influxdb2.Client
	measurement string
	bucket      string
	org         string
	tags        map[string]string
	fields      []string
	logr        *log.Logger
}

func NewReporter(host, token, org, bucket, measurement string, tags map[string]string, fields []string, logr *log.Logger) database.Interface {
	client := influxdb2.NewClient(host, token)
	return &Influxdb2Reporter{
		org:         org,
		logr:        logr,
		tags:        tags,
		client:      client,
		fields:      fields,
		bucket:      bucket,
		measurement: measurement,
	}
}

func (r *Influxdb2Reporter) Close() {
	r.client.Close()
}

func (r *Influxdb2Reporter) Report(batch []*weather.Weather) error {
	writer := r.client.WriteAPIBlocking(r.org, r.bucket)
	batchedPoints := []*write.Point{}

	for _, w := range batch {
		p := influxdb2.NewPoint(
			r.measurement, r.tags, toFields(r.fields, w.Measurements), w.Time,
		)
		batchedPoints = append(batchedPoints, p)
	}
	err := writer.WritePoint(context.Background(), batchedPoints...)
	if err != nil {
		r.logr.Println("influxdb2 failed to WritePoint", err)
	}
	return err
}

func toFields(targetFields []string, weatherfeilds []weather.Measurement) map[string]interface{} {
	fields := make(map[string]interface{})
	for _, field := range targetFields {
		fields[field] = getValue(field, weatherfeilds)
	}
	return fields
}

func getValue(field string, weatherfeilds []weather.Measurement) interface{} {
	for _, wfield := range weatherfeilds {
		if wfield.Name == field {
			return wfield.Value
		}
	}
	return nil
}
