package main

import (
	"flag"
	"log"
	"time"

	"github.com/jpxor/go-weather-reporter/internal"
	"github.com/jpxor/ssconfig"
)

var DEFAULT_INTERVAL = time.Minute * 10

func main() {
	logr := log.New(log.Writer(), "go-weather-reporter: ", log.LstdFlags|log.Lmsgprefix)
	config := parseArgs(logr)
	internal.Run(config, logr)
}

func parseArgs(logr *log.Logger) internal.Config {
	logr.Println("parsing inputs")

	opts := struct {
		ConfigPath     string
		QueryInterval  int
		ReportInterval int
		Start          bool
	}{}

	flag.StringVar(&opts.ConfigPath, "config", "", "Set path to a configuration file")
	flag.BoolVar(&opts.Start, "start", false, "Tells weather-reporter to start, continues until stopped")

	// the following flags overwrite values set in the config file,
	// the default value must be invalid so we know when to NOT overwrite the config
	// the real default values are set above
	flag.IntVar(&opts.QueryInterval, "interval", 0, "Query interval in seconds")
	flag.IntVar(&opts.ReportInterval, "report_interval", 0, "Set report interval in seconds if it should differ from the query interval")

	flag.Parse()

	config := internal.Config{}
	if opts.ConfigPath != "" {
		ssconfig.Set{FilePath: opts.ConfigPath}.Load(&config)
	}

	// check for valid interval,
	// then check if default value is needed
	if opts.QueryInterval > 0 {
		config.QueryInterval = time.Duration(opts.QueryInterval)
	} else if config.QueryInterval == 0 {
		config.QueryInterval = DEFAULT_INTERVAL
	}

	// check for valid interval,
	// then check if default value is needed
	if opts.ReportInterval > 0 {
		config.ReportInterval = time.Duration(opts.ReportInterval)
	} else if config.ReportInterval == 0 {
		config.ReportInterval = config.QueryInterval
	}

	config.Start = opts.Start
	return config
}
