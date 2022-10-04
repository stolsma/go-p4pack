// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package dpdkinfra

/*
#cgo pkg-config: libdpdk
*/
import "C"

import (
	"errors"

	"github.com/stolsma/go-p4pack/pkg/dpdkswx"
)

// RingStore represents a store of created Tap records
type RingStore map[string]*dpdkswx.Ring

func CreateRingStore() RingStore {
	return make(RingStore)
}

func (rs RingStore) Find(name string) *dpdkswx.Ring {
	if name == "" {
		return nil
	}

	return rs[name]
}

// Create Ring interface. Returns a pointer to a Ring structure or nil with error.
func (rs RingStore) Create(name string, size uint32, numaNode uint32) (*dpdkswx.Ring, error) {
	var ring dpdkswx.Ring

	if rs.Find(name) != nil {
		return nil, errors.New("ring interface exists")
	}

	// create callback function called when record is freed
	clean := func() {
		delete(rs, name)
	}

	// initialize
	ring.Init(name, size, numaNode, clean)

	// add node to list
	rs[name] = &ring

	return &ring, nil
}

// Delete all Ring interfaces
func (rs RingStore) Clear() {
	for _, ring := range rs {
		ring.Free()
	}
}
