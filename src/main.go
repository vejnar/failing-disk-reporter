//
// Copyright (C) 2020 Charles E. Vejnar
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://www.mozilla.org/MPL/2.0/.
//

package main

import (
	"flag"
	"log"
)

func main() {
	// Arguments
	var configPath string
	var forceReport, debug, verbose bool
	flag.StringVar(&configPath, "config", "fdr.toml", "Config path")
	flag.BoolVar(&forceReport, "report", false, "Send reports ignoring intervals")
	flag.BoolVar(&debug, "debug", false, "Debug")
	flag.BoolVar(&verbose, "verbose", false, "Verbose")
	// Arguments: Parse
	flag.Parse()

	var err error

	// Open config
	var config *Config
	config, err = ParseConfig(configPath)
	if err != nil {
		log.Fatal(err)
	}
	// Verbose
	if verbose {
		config.Verbose = true
	}
	if debug {
		config.Debug = true
		config.Verbose = true
	}

	// Get devices
	var devices *Devices
	devices, err = NewDevices(config.IgnoredProtocols, config.Debug)
	if err != nil {
		log.Fatal(err)
	}
	if config.Verbose {
		log.Printf("%d device(s) detected", devices.Length())
	}
	if config.Debug {
		for i, d := range *devices {
			log.Printf("Device %2d: %-20s %-20s %s", i+1, d.Type, d.Name, d.Protocol)
		}
	}

	// Find error(s)
	err = devices.FindErrors(config.Criteria, config.Debug)
	if err != nil {
		log.Fatal(err)
	}
	devices.RemoveDuplicates()
	if config.Verbose {
		log.Printf("%d error(s) detected", devices.LengthError())
	}
	if config.Debug && devices.LengthError() > 0 {
		r := TextReporter{name: "text"}
		r.Report(devices, forceReport)
	}

	// Report
	var sent bool
	for _, r := range config.Reporters {
		sent, err = r.Report(devices, forceReport)
		if err != nil {
			log.Fatal(err)
		}
		if config.Verbose && sent {
			log.Printf("Report sent to %s", r.Name())
		}
	}
}
