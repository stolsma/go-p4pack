// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package dpdkinfra

/*
#cgo pkg-config: libdpdk
*/
import "C"

import (
	"errors"
)

// TapStore represents a store of created Tap records
type TapStore map[string]*Tap

func CreateTapStore() TapStore {
	return make(TapStore)
}

// Find tap record with name
func (ts TapStore) Find(name string) *Tap {
	if name == "" {
		return nil
	}

	return ts[name]
}

// Create Tap interface. Returns a pointer to a Tap structure or nil with error.
func (ts TapStore) Create(name string, tapConfig *TapConfig) (*Tap, error) {
	tap := Tap{}

	if ts.Find(name) != nil {
		return nil, errors.New("tap interface exists")
	}

	// create callback function called when record is freed
	clean := func() {
		delete(ts, name)
	}

	// initialize
	tap.Init(name, tapConfig, clean)

	// add node to list
	ts[name] = &tap

	return &tap, nil
}

// Delete all Tap interfaces and close corresponding fd
func (ts TapStore) Clear() {
	for _, tap := range ts {
		tap.Free()
	}
}

func (ts TapStore) Iterate(fn func(key string, tap *Tap) error) error {
	if fn != nil {
		for k, v := range ts {
			if err := fn(k, v); err != nil {
				return err
			}
		}
	} else {
		return errors.New("no function to call")
	}
	return nil
}
