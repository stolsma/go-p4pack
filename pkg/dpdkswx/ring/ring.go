// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package ring

/*
#include <rte_ring.h>

*/
import "C"
import (
	"errors"
	"unsafe"

	"github.com/stolsma/go-p4pack/pkg/dpdkswx/common"
)

const (
	// SingleConsumer specifies that default dequeue operation will exhibit 'single-consumer' behaviour.
	SingleConsumer uint = C.RING_F_SC_DEQ
	// SingleProducer specifies that default enqueue operation will exhibit 'single-producer' behaviour.
	SingleProducer = C.RING_F_SP_ENQ
	// ExactSize specifies how to handle ring size during Create/Init. Ring is to hold exactly requested number of
	// entries. Without this flag set, the ring size requested must be a power of 2, and the usable space will be that
	// size - 1. With the flag, the requested size will be rounded up to the next power of two, but the usable space will
	// be exactly that requested. Worst case, if a power-of-2 size is requested, half the ring space will be wasted.
	ExactSize = C.RING_F_EXACT_SZ
)

// Ring represents a Ring record
type Ring struct {
	name  string
	r     *C.struct_rte_ring
	size  uint
	clean func()
}

// Create Ring interface. Returns a pointer to a Ring structure or nil with error.
func (r *Ring) Init(name string, size uint, numaNode uint32, clean func()) error {
	const flags = SingleProducer | SingleConsumer

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
	r.clean = clean

	return nil
}

func (r *Ring) Name() string {
	return r.name
}

func (r *Ring) Ring() unsafe.Pointer {
	return unsafe.Pointer(r.r)
}

func (r *Ring) Size() uint {
	return r.size
}

// Free deletes the current Ring record and calls the clean callback function given at init
func (r *Ring) Free() {
	C.rte_ring_free(r.r)

	// call given clean callback function if given during init
	if r.clean != nil {
		r.clean()
	}
}

// Lookup searches a ring from its name in RTE_TAILQ_RING, i.e. among those created with rte_ring_create.
// Returns a partly initialized ring object
func Lookup(name string, clean func()) (*Ring, error) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	cr := C.rte_ring_lookup(cname)
	if cr == nil {
		return nil, common.Err()
	}

	size := uint(C.rte_ring_get_size(cr))

	// Node fill in
	r := &Ring{
		name:  name,
		r:     cr,
		size:  size,
		clean: clean,
	}

	return r, nil
}
