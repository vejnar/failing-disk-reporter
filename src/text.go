//
// Copyright (C) 2020 Charles E. Vejnar
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://www.mozilla.org/MPL/2.0/.
//

package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/pelletier/go-toml"
)

type TextReporter struct {
	name        string
	errorHeader string
	errorFooter string
}

func NewTextReporter(t *toml.Tree) (*TextReporter, error) {
	r := TextReporter{}
	// Get strings
	r.name = t.Get("name").(string)
	r.errorHeader = t.Get("error_header").(string)
	r.errorFooter = t.Get("error_footer").(string)
	return &r, nil
}

func (r *TextReporter) Name() string {
	return r.name
}

func (r *TextReporter) Report(devices *Devices, forceReport bool) (sendMsg bool, err error) {
	var hostname, msg, msgHeader, msgFooter string

	// Hostname
	hostname, err = os.Hostname()
	if err != nil {
		return false, err
	}
	// Now
	t := time.Now()

	// Prepare main message
	if devices.LengthError() > 0 {
		// Error message
		msgHeader = r.errorHeader
		msgFooter = r.errorFooter
		msg = "Failing drives\n"
		for _, device := range *devices {
			if len(device.Errors) > 0 {
				msg += fmt.Sprintf("* %s (%s) on %s\n", device.Model, device.SerialNumber, device.Name)
				for _, deviceError := range device.Errors {
					msg += fmt.Sprintf("Error: %s at %d (max:%d)\n", deviceError.Criteria.Label, deviceError.Value, deviceError.Criteria.Max)
				}
			}
		}
	}

	// Write message
	if len(msgHeader) > 0 {
		msgHeader = msgHeader + "\n"
	}
	if len(msgFooter) > 0 {
		msgFooter = "\n" + msgFooter
	}
	log.Printf("%s%s on %s\n%s%s", msgHeader, t.Format("2006-01-02 15:04"), hostname, msg, msgFooter)

	return true, nil
}
