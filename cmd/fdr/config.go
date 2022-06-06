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

	"github.com/pelletier/go-toml"
)

type Criteria struct {
	Protocol string
	Key      string
	ID       int
	Name     string
	Label    string
	Max      int
}

type Config struct {
	IgnoredProtocols map[string]bool
	Criteria         map[string][]Criteria
	Reporters        []Reporter
	Verbose          bool
	Debug            bool
}

func ParseConfig(path string) (*Config, error) {
	config := &Config{}
	var err error

	// Open config
	var tree *toml.Tree
	tree, err = toml.LoadFile(path)
	if err != nil {
		return config, err
	}

	// Verbose
	v := tree.Get("general.verbose")
	if v == nil {
		config.Verbose = false
	} else {
		config.Verbose = v.(bool)
	}

	// Ignored protocols
	config.IgnoredProtocols = make(map[string]bool)
	k := tree.Get("smart.ignored_protocols")
	if k != nil {
		for _, p := range k.([]interface{}) {
			config.IgnoredProtocols[p.(string)] = true
		}
	}

	// SMART Criteria
	config.Criteria = make(map[string][]Criteria)
	for _, t := range tree.Get("smart.criteria").([]*toml.Tree) {
		c := Criteria{}
		t.Unmarshal(&c)
		config.Criteria[c.Protocol] = append(config.Criteria[c.Protocol], c)
	}

	// Reporters
	for _, t := range tree.Get("reporters").([]*toml.Tree) {
		var name string
		raw := t.Get("name")
		if raw == nil {
			return config, fmt.Errorf("Config: Missing name in reporter")
		} else {
			var r Reporter
			name = raw.(string)
			switch name {
			case "matrix":
				r, err = NewMatrixReporter(t)
			case "slack":
				r, err = NewSlackReporter(t)
			default:
				return config, fmt.Errorf("Unknown report name: %s", name)
			}
			if err != nil {
				return config, err
			}
			config.Reporters = append(config.Reporters, r)
		}
	}

	return config, err
}
