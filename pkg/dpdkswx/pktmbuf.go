// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package dpdkswx

/*
#cgo pkg-config: libdpdk

#include <rte_mempool.h>
#include <rte_mbuf.h>

*/
import "C"
import (
	"errors"
	"unsafe"
)

var BufferSizeMin int = C.sizeof_struct_rte_mbuf + C.RTE_PKTMBUF_HEADROOM

// Pktmbuf represents a DPDK Packet memory buffer (pktmbuf) stores in a Pktmbuf store
type Pktmbuf struct {
	name       string
	m          *C.struct_rte_mempool
	bufferSize uint32
	clean      func()
}

// Create pktmbuf with corresponding pktmbuf mempool memory. Returns a pointer to a Pktmbuf
// structure or nil with error.
func (pm *Pktmbuf) Init(
	name string,
	bufferSize uint32,
	poolSize uint32,
	cacheSize uint32,
	cpuID int,
	clean func()) error {
	var m *C.struct_rte_mempool

	// Check input params
	if name == "" {
		return errors.New("no name given")
	}

	if bufferSize < (uint32)(BufferSizeMin) {
		return errors.New("buffer size to small")
	}

	if poolSize == 0 {
		return errors.New("pool size is 0")
	}

	// Resource create
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	m = C.rte_pktmbuf_pool_create(
		cname,
		C.uint(poolSize),
		C.uint(cacheSize),
		C.ushort(0),
		C.ushort(bufferSize-C.sizeof_struct_rte_mbuf),
		C.int(cpuID),
	)
	if m == nil {
		return err()
	}

	// Node fill in
	pm.name = name
	pm.m = m
	pm.bufferSize = bufferSize
	pm.clean = clean

	return nil
}

func (pm *Pktmbuf) Name() string {
	return pm.name
}

func (pm *Pktmbuf) Mempool() *C.struct_rte_mempool {
	return pm.m
}

func (pm *Pktmbuf) Free() {
	if pm.m != nil {
		C.rte_mempool_free(pm.m)
		pm.m = nil
	}

	// call given clean callback function if given during init
	if pm.clean != nil {
		pm.clean()
	}
}
