//
// Copyright (C) 2020 Charles E. Vejnar
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://www.mozilla.org/MPL/2.0/.
//

package main

import (
	"os"
	"time"
)

type Reporter interface {
	Name() string
	Report(*Devices) (bool, error)
}

func touchFile(path string) error {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		file, err := os.Create(path)
		if err != nil {
			return err
		}
		defer file.Close()
	} else {
		currentTime := time.Now().Local()
		err = os.Chtimes(path, currentTime, currentTime)
		if err != nil {
			return err
		}
	}
	return nil
}

func decideReport(path string, interval time.Duration) (sendMsg bool, err error) {
	// Report by default
	if path == "" {
		return true, nil
	}
	var stat os.FileInfo
	// Flag file not found
	if stat, err = os.Stat(path); os.IsNotExist(err) {
		err = touchFile(path)
		if err != nil {
			return
		}
		sendMsg = true
	} else {
		t := time.Now()
		// Flag file too old
		if t.After(stat.ModTime().Add(interval)) {
			err = touchFile(path)
			if err != nil {
				return
			}
			sendMsg = true
		}
	}
	return
}
