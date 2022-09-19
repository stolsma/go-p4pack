// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package dpdkinfra

/*
#cgo pkg-config: libdpdk
*/
import "C"

import (
	"errors"

	"github.com/stolsma/go-p4dpdk-vswitch/pkg/dpdkswx"
)

// TapStore represents a store of created Tap records
type TapStore map[string]*dpdkswx.Tap

func CreateTapStore() TapStore {
	return make(TapStore)
}

func (ts TapStore) Find(name string) *dpdkswx.Tap {
	if name == "" {
		return nil
	}

	return ts[name]
}

// Create Tap interface. Returns a pointer to a Tap structure or nil with error.
func (ts TapStore) Create(name string) (*dpdkswx.Tap, error) {
	var tap dpdkswx.Tap

	if ts.Find(name) != nil {
		return nil, errors.New("tap interface exists")
	}

	// create callback function called when record is freed
	clean := func() {
		delete(ts, name)
	}

	// initialize
	tap.Init(name, clean)

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
