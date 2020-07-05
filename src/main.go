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
	"fmt"
	"log"
)

func main() {
	// Arguments
	var configPath string
	flag.StringVar(&configPath, "config", "fdr.toml", "Config path")
	// Arguments: Parse
	flag.Parse()

	var err error

	// Open config
	var config *Config
	config, err = ParseConfig(configPath)
	if err != nil {
		log.Fatal(err)
	}

	// Get devices
	var devices *Devices
	devices, err = NewDevices(config.IgnoredProtocols)
	if err != nil {
		log.Fatal(err)
	}

	// Find error(s)
	err = devices.FindErrors(config.Criteria)
	if err != nil {
		log.Fatal(err)
	}
	devices.RemoveDuplicates()
	fmt.Printf("%d devices and %d errors detected\n", devices.Length(), devices.LengthError())

	// Report
	var sent bool
	for _, r := range config.Reporters {
		sent, err = r.Report(devices)
		if err != nil {
			log.Fatal(err)
		}
		if sent {
			fmt.Printf("Report sent to %s\n", r.Name())
		}
	}
}
