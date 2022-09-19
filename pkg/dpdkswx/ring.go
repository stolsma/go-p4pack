// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package dpdkswx

/*
#cgo pkg-config: libdpdk

#include <rte_ring.h>

*/
import "C"
import (
	"errors"
	"unsafe"
)

// Ring represents a Ring record
type Ring struct {
	name     string
	r        *C.struct_rte_ring
	size     uint32
	numaNode uint32
	clean    func()
}

// Create Ring interface. Returns a pointer to a Ring structure or nil with error.
func (r *Ring) Init(name string, size uint32, numaNode uint32, clean func()) error {
	const flags int = C.RING_F_SP_ENQ | C.RING_F_SC_DEQ

	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	// Resource create
	ring := C.rte_ring_create(cname, (C.uint)(size), (C.int)(numaNode), (C.uint)(flags))
	if r == nil {
		return errors.New("Ring creation error")
	}

	// Node fill in
	r.name = name
	r.r = ring
	r.size = size
	r.numaNode = numaNode
	r.clean = clean

	return nil
}

// Free deletes the current Ring record and calls the clean callback function given at init
func (r *Ring) Free() {
	C.rte_ring_free(r.r)

	// call given clean callback function if given during init
	if r.clean != nil {
		r.clean()
	}
}
