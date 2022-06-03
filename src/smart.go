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
	"log"
	"os/exec"

	"github.com/buger/jsonparser"
)

type Cat struct {
	Ata  int
	Nvme int
	P    string
}

type Smart struct {
	Ata  Cat
	Nvme Cat
}

type Device struct {
	Type         string
	Name         string
	Protocol     string
	Model        string
	SerialNumber string
	Duplicate    bool
	Errors       []DeviceError
}

type DeviceError struct {
	Criteria
	Value int
}

type Devices []Device

func parseScan(data []byte, ignoredProtocols map[string]bool) (*Devices, error) {
	var devices Devices
	var err error
	jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err2 error) {
		if err == nil {
			var device Device
			device.Type, err = jsonparser.GetString(value, "type")
			if err != nil {
				return
			}
			device.Name, err = jsonparser.GetString(value, "name")
			if err != nil {
				return
			}
			device.Protocol, err = jsonparser.GetString(value, "protocol")
			if err != nil {
				return
			}
			// Ignored protocols
			if _, ok := ignoredProtocols[device.Protocol]; !ok {
				devices = append(devices, device)
			}
		}
	}, "devices")
	return &devices, err
}

func parseDeviceInfo(data []byte, device *Device) error {
	m, err := jsonparser.GetString(data, "model_name")
	if err != nil {
		return err
	} else {
		device.Model = m
	}
	sn, err := jsonparser.GetString(data, "serial_number")
	if err != nil {
		return err
	} else {
		device.SerialNumber = sn
	}
	return err
}

func parseError(data []byte) (errorMessage string, err error) {
	jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err2 error) {
		errorMessage, err = jsonparser.GetString(value, "string")
		if err != nil {
			return
		}
	}, "smartctl", "messages")
	return errorMessage, err
}

func switchSmartError(ecode int, out []byte, deviceName string, debug bool) error {
	// Ignore return codes 4, 64, 68 (might indicate some past failure; mostly false positive)
	switch ecode {
	case 4:
		if debug {
			log.Printf("smartctl returned with code %d on %s", ecode, deviceName)
		}
	case 64:
		if debug {
			log.Printf("smartctl returned with code %d on %s", ecode, deviceName)
		}
	case 68:
		if debug {
			log.Printf("smartctl returned with code %d on %s", ecode, deviceName)
		}
	default:
		errorMessage, err := parseError(out)
		if err != nil {
			return err
		} else {
			return fmt.Errorf("smartctl failed on %s (code %d): %s", deviceName, ecode, errorMessage)
		}
	}
	return nil
}

func hasKey(data []byte, keys ...string) bool {
	_, _, _, err := jsonparser.Get(data, keys...)
	if err == jsonparser.KeyPathNotFoundError {
		return false
	}
	return true
}

func NewDevices(ignoredProtocols map[string]bool, debug bool) (devices *Devices, err error) {
	var out []byte

	// Scan devices
	if debug {
		log.Println("Start smartctl --scan-open")
		out, err = exec.Command("smartctl", "--scan-open").Output()
		if err != nil {
			return devices, err
		}
		log.Print(string(out))
	}
	out, err = exec.Command("smartctl", "--scan-open", "--json").Output()
	if err != nil {
		return devices, err
	}
	// Parse
	devices, err = parseScan(out, ignoredProtocols)

	return devices, err
}

func (devices *Devices) Length() int {
	return len(*devices)
}

func (devices *Devices) LengthError() (n int) {
	for _, device := range *devices {
		n += len(device.Errors)
	}
	return
}

func (devices *Devices) FindErrors(criteria map[string][]Criteria, debug bool) (err error) {
	var out []byte

	for idevice := 0; idevice < devices.Length(); idevice++ {
		device := &(*devices)[idevice]

		// Run smartctl
		if debug {
			log.Println("Start smartctl --all --device", device.Type, device.Name)
			out, err = exec.Command("smartctl", "--all", "--device", device.Type, device.Name).Output()
			if err != nil {
				if exitError, ok := err.(*exec.ExitError); ok {
					switchSmartError(exitError.ExitCode(), out, device.Name, debug)
				}
			}
			log.Print(string(out))
		}
		out, err = exec.Command("smartctl", "--all", "--json", "--device", device.Type, device.Name).Output()
		if err != nil {
			if exitError, ok := err.(*exec.ExitError); ok {
				switchSmartError(exitError.ExitCode(), out, device.Name, debug)
			}
		}

		// Get device info
		err = parseDeviceInfo(out, device)
		if err != nil {
			return err
		}

		// New device?
		dupDevice := false
		for id, d := range *devices {
			if id != idevice && d.Model == device.Model && d.SerialNumber == device.SerialNumber {
				dupDevice = true
				break
			}
		}
		// Skip duplicate device
		if dupDevice {
			device.Duplicate = true
			continue
		}

		if cts, ok := criteria[device.Protocol]; ok {
			var nFound int
			for _, ct := range cts {
				var cValue int
				// If key defines the name of a table (otherwise get key directly)
				if hasKey(out, ct.Key, "table") {
					jsonparser.ArrayEach(out, func(value []byte, dataType jsonparser.ValueType, offset int, err2 error) {
						if err == nil {
							// Match key by ID or name
							var k1 int64
							var k2 string
							k1, err = jsonparser.GetInt(value, "id")
							if err == jsonparser.KeyPathNotFoundError {
								k1 = -1
							} else if err != nil {
								return
							}
							k2, err = jsonparser.GetString(value, "name")
							if err == jsonparser.KeyPathNotFoundError {
								k2 = "xyz"
							} else if err != nil {
								return
							}
							if int(k1) == ct.ID || k2 == ct.Name {
								var v int64
								v, err = jsonparser.GetInt(value, "raw", "value")
								if err != nil {
									return
								}
								cValue = int(v)
								nFound++
							}
						}
					}, ct.Key, "table")
					if err != nil {
						return err
					}
				} else {
					v, err := jsonparser.GetInt(out, ct.Key, ct.Name)
					if err != nil {
						return err
					}
					cValue = int(v)
					nFound++
				}
				// Check value
				if ct.Max < cValue {
					device.Errors = append(device.Errors, DeviceError{Criteria: ct, Value: cValue})
				}
			}
			if nFound < 1 {
				return fmt.Errorf("No info about drive")
			}
		} else {
			return fmt.Errorf("Protocol %s not found", device.Protocol)
		}
	}

	return err
}

func (devices *Devices) RemoveDuplicates() {
	var newDevices Devices
	for _, device := range *devices {
		if !device.Duplicate {
			newDevices = append(newDevices, device)
		}
	}
	*devices = newDevices
}
