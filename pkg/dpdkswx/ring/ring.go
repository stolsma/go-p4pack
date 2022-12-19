// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package ring

/*
#include <rte_ring.h>
#include <rte_swx_port_ring.h>

*/
import "C"
import (
	"errors"
	"unsafe"

	"github.com/stolsma/go-p4pack/pkg/dpdkswx/common"
	"github.com/stolsma/go-p4pack/pkg/dpdkswx/device"
	"github.com/stolsma/go-p4pack/pkg/dpdkswx/pipeline"
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

type Params struct {
	Size     uint
	NumaNode uint32
}

// Ring represents a Ring device
type Ring struct {
	*device.Device
	r    *C.struct_rte_ring
	size uint // must be a power of two
}

// Create and initialize Ring device
func (r *Ring) Init(name string, params *Params, clean func()) error {
	const flags = SingleProducer | SingleConsumer

	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	// Resource create
	ring := C.rte_ring_create(cname, (C.uint)(params.Size), (C.int)(params.NumaNode), (C.uint)(flags))
	if r == nil {
		return errors.New("Ring creation error")
	}

	// Node fill in
	r.Device = &device.Device{}
	r.SetType("RING")
	r.SetName(name)
	r.r = ring
	r.size = params.Size

	r.SetPipelineOutPort(device.NotBound)
	r.SetPipelineInPort(device.NotBound)
	r.SetClean(clean)

	return nil
}

func (r *Ring) Ring() unsafe.Pointer {
	return unsafe.Pointer(r.r)
}

func (r *Ring) Size() uint {
	return r.size
}

// Free deletes the current Ring record and calls the clean callback function given at init
func (r *Ring) Free() error {
	C.rte_ring_free(r.r)

	// call given clean callback function if given during init
	if r.Clean() != nil {
		r.Clean()()
	}

	return nil
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
		Device: &device.Device{},
		r:      cr,
		size:   size,
	}
	r.SetType("RING")
	r.SetName(name)
	r.SetPipelineOutPort(device.NotBound)
	r.SetPipelineInPort(device.NotBound)
	r.SetClean(clean)

	return r, nil
}

// bind to given pipeline input port
func (r *Ring) BindToPipelineInputPort(pl *pipeline.Pipeline, portID int, rxq uint, bsz uint) error {
	var params C.struct_rte_swx_port_ring_reader_params

	if r.PipelineInPort() != device.NotBound {
		return errors.New("port already bound")
	}
	r.SetPipelineIn(pl.GetName())
	r.SetPipelineInPort(portID)

	params.name = C.CString(r.Name())
	defer C.free(unsafe.Pointer(params.name))
	params.burst_size = (C.uint)(bsz)

	return pl.PortInConfig(portID, "ring", unsafe.Pointer(&params))
}

// bind to given pipeline output port
func (r *Ring) BindToPipelineOutputPort(pl *pipeline.Pipeline, portID int, txq uint, bsz uint) error {
	var params C.struct_rte_swx_port_ring_writer_params

	if r.PipelineOutPort() != device.NotBound {
		return errors.New("port already bound")
	}
	r.SetPipelineOut(pl.GetName())
	r.SetPipelineOutPort(portID)

	params.name = C.CString(r.Name())
	defer C.free(unsafe.Pointer(params.name))
	params.burst_size = (C.uint)(bsz)

	return pl.PortOutConfig(portID, "ring", unsafe.Pointer(&params))
}
