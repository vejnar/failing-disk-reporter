//
// Copyright (C) 2020-2022 Charles E. Vejnar
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://www.mozilla.org/MPL/2.0/.
//

package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/pelletier/go-toml"
)

type MatrixReporter struct {
	name          string
	url           string
	errorHeader   string
	errorFooter   string
	okInterval    time.Duration
	okPath        string
	errorInterval time.Duration
	errorPath     string
}

func NewMatrixReporter(t *toml.Tree) (*MatrixReporter, error) {
	var err error
	r := MatrixReporter{}
	// Get strings
	r.name = t.Get("name").(string)
	r.url = t.Get("url").(string)
	r.errorHeader = t.Get("error_header").(string)
	r.errorFooter = t.Get("error_footer").(string)
	r.okPath = t.Get("ok_path").(string)
	r.errorPath = t.Get("error_path").(string)
	// Parse durations
	var d time.Duration
	d, err = time.ParseDuration(t.Get("ok_interval").(string))
	if err != nil {
		return &r, err
	} else {
		r.okInterval = d
	}
	d, err = time.ParseDuration(t.Get("error_interval").(string))
	if err != nil {
		return &r, err
	} else {
		r.errorInterval = d
	}
	return &r, nil
}

func (r *MatrixReporter) Name() string {
	return r.name
}

func (r *MatrixReporter) Report(devices *Devices, forceReport bool) (sendMsg bool, err error) {
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
		msg = "Failing drives\\n"
		for _, device := range *devices {
			if len(device.Errors) > 0 {
				msg += fmt.Sprintf("* %s (%s) on %s\\n", device.Model, device.SerialNumber, device.Name)
				for _, deviceError := range device.Errors {
					msg += fmt.Sprintf("\\tError: %s at %d (max:%d)\\n", deviceError.Criteria.Label, deviceError.Value, deviceError.Criteria.Max)
				}
			}
		}
		if forceReport {
			sendMsg = true
		} else {
			sendMsg, err = decideReport(r.errorPath, r.errorInterval)
			if err != nil {
				return false, err
			}
		}
	} else {
		// OK message
		msg = fmt.Sprintf("%d drives found\\n", devices.Length())
		if forceReport {
			sendMsg = true
		} else {
			sendMsg, err = decideReport(r.okPath, r.okInterval)
			if err != nil {
				return false, err
			}
		}
	}

	// Send message to Matrix
	if sendMsg {
		if len(msgHeader) > 0 {
			msgHeader = msgHeader + "\\n\\n"
		}
		if len(msgFooter) > 0 {
			msgFooter = "\\n" + msgFooter
		}
		jsonString := fmt.Sprintf("{\"msgtype\":\"m.text\", \"body\":\"%s%s on %s\\n%s%s\"}", msgHeader, t.Format("2006-01-02 15:04"), hostname, msg, msgFooter)

		// Send message
		resp, err := http.Post(r.url, "application/json", strings.NewReader(jsonString))
		defer resp.Body.Close()

		switch {
		case err != nil:
			return false, fmt.Errorf("Network error: %s", err)
		case resp.StatusCode != 200:
			return false, fmt.Errorf("Bad HTTP status code: %d", resp.StatusCode)
		}
	}

	return sendMsg, nil
}
