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
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/jpxor/go-weather-reporter/integrations"
	"github.com/jpxor/go-weather-reporter/integrations/weather/openweathermap"
)

func getDuration(str, suffix string, scale time.Duration) (time.Duration, bool) {
	if strings.HasSuffix(str, suffix) {
		str = strings.TrimSuffix(str, suffix)
		interval, err := strconv.Atoi(str)
		if err != nil {
			fmt.Println(err)
			return 0, false
		}
		return scale * time.Duration(interval), true
	}
	return 0, false
}

func getPollInterval(val interface{}) (time.Duration, bool) {
	str, ok := val.(string)
	if !ok {
		return 0, false
	}
	interval, ok := getDuration(str, "s", time.Second)
	if ok {
		return interval, true
	}
	interval, ok = getDuration(str, "m", time.Minute)
	if ok {
		return interval, true
	}
	interval, ok = getDuration(str, "h", time.Hour)
	if ok {
		return interval, true
	}
	interval, ok = getDuration(str, "d", 24*time.Hour)
	if ok {
		return interval, true
	}
	// suffix unknown or none
	// TODO: test if unknown suffix is used
	fmt.Println("warning: unknown duration format, defaulting to seconds (s)")
	return getDuration(str, "", time.Second)
}

func getSourceIntegration(name string) integrations.SourceInterface {
	switch name {
	case "openweathermap":
		return &openweathermap.OpenWeatherService{}
	}
	return nil
}

func StartService(interval time.Duration, source integrations.SourceInterface, dests []integrations.DestinationInterface) {
	fmt.Println("started!")
}

func Run(config Config, opts Opts, logr *log.Logger) {

	for _, service := range config {
		logr.Println("Starting service:", service.Name)

		sourceName, ok := service.Source["name"].(string)
		if !ok {
			logr.Fatalln("service is missing a name, config:", service.ConfPath)
		}
		poll_interval, ok := getPollInterval(service.Source["poll_interval"])
		if !ok {
			logr.Fatalln("failed to parse poll_interval, config:", service.ConfPath)
		}
		source := getSourceIntegration(sourceName)
		if source == nil {
			logr.Fatalln("no source integration with name:", sourceName)
		}
		source.Init(service.Source)

		var dests []integrations.DestinationInterface
		go StartService(poll_interval, source, dests)

		// TEMP
		time.Sleep(time.Minute)
	}
}
