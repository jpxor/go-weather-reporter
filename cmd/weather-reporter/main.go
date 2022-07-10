package main

import (
	"log"

	"github.com/jpxor/go-weather-reporter/internal"
)

func main() {
	logger := log.Default()
	opts := parseArgs()
	internal.Run(opts, logger)
}

func parseArgs() internal.Opts {
	return internal.Opts{}
}
