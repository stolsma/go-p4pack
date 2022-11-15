// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package dpdkinfra

/*
#cgo pkg-config: libdpdk
*/
import "C"

import (
	"errors"

	"github.com/stolsma/go-p4pack/pkg/dpdkswx/ring"
)

// RingStore represents a store of created Tap records
type RingStore map[string]*ring.Ring

func CreateRingStore() RingStore {
	return make(RingStore)
}

func (rs RingStore) Find(name string) *ring.Ring {
	if name == "" {
		return nil
	}

	return rs[name]
}

// Create Ring interface. Returns a pointer to a Ring structure or nil with error.
func (rs RingStore) Create(name string, size uint, numaNode uint32) (*ring.Ring, error) {
	var r ring.Ring

	if rs.Find(name) != nil {
		return nil, errors.New("ring interface exists")
	}

	// create callback function called when record is freed
	clean := func() {
		delete(rs, name)
	}

	// initialize
	r.Init(name, size, numaNode, clean)

	// add node to list
	rs[name] = &r

	return &r, nil
}

// Delete all Ring interfaces
func (rs RingStore) Clear() {
	for _, r := range rs {
		r.Free()
	}
}
