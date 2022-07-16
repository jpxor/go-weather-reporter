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

package main

import (
	"flag"
	"log"

	"github.com/jpxor/go-weather-reporter/internal"
)

func main() {
	logr := log.New(log.Writer(), "go-data-logger: ", log.LstdFlags|log.Lmsgprefix)
	config, opts := parseArgs(logr)
	internal.Run(config, opts, logr)
}

func parseArgs(logr *log.Logger) (internal.Config, internal.Opts) {
	opts := internal.Opts{}

	flag.StringVar(&opts.ConfigDir, "cdir", "./config/", "Set path to a directory containing config files")
	flag.BoolVar(&opts.Once, "once", false, "Execute each query once, then exit")
	flag.Parse()

	logr.Println("parsing config files")
	parser := internal.NewConfigParser(logr)
	config, err := parser.ParseConfigFiles(opts.ConfigDir)

	if err != nil {
		logr.Println(err)
		logr.Fatalln("faild to parse config files")
	}

	return config, opts
}
