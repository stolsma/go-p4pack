// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package dpdkinfra

import (
	"errors"

	"github.com/stolsma/go-p4pack/pkg/dpdkswx"
	"github.com/stolsma/go-p4pack/pkg/dpdkswx/pktmbuf"
	"github.com/stolsma/go-p4pack/pkg/dpdkswx/swxruntime"
	"github.com/yerden/go-dpdk/mempool"
)

var BufferSizeMin uint = pktmbuf.SizeofRteMbuf + pktmbuf.RtePktmbufHeadroom

// Pktmbuf represents a DPDK Packet memory buffer (pktmbuf) initialized memory pool (mempool)
type Pktmbuf struct {
	name       string
	m          *mempool.Mempool
	bufferSize uint
	clean      func()
}

// Create pktmbuf with corresponding pktmbuf mempool memory. Returns a pointer to a Pktmbuf
// structure or nil with error.
func (pm *Pktmbuf) Init(
	name string,
	bufferSize uint,
	poolSize uint32,
	cacheSize uint32,
	cpuSocketID int,
	clean func(),
) error {
	// Check input params
	if name == "" {
		return errors.New("no name given")
	}

	if bufferSize < BufferSizeMin {
		return errors.New("buffer size to small")
	}

	if poolSize == 0 {
		return errors.New("pool size is 0")
	}

	// create PktMbufpool on main DPDK Lcore to prevent problems
	var mpErr error
	var md *mempool.Mempool
	if execErr := dpdkswx.Runtime.ExecOnMain(func(*swxruntime.MainCtx) {
		md, mpErr = mempool.CreateMbufPool(
			name,
			poolSize,
			uint16(bufferSize),
			mempool.OptSocket(cpuSocketID),
			mempool.OptCacheSize(cacheSize),
			mempool.OptPrivateDataSize(0), // for each Mbuf
		)
	}); execErr != nil {
		return execErr
	}

	// error during CreateMbufPool ?
	if mpErr != nil {
		return mpErr
	}

	// Node fill in
	pm.name = name
	pm.m = md
	pm.bufferSize = bufferSize
	pm.clean = clean

	return nil
}

func (pm *Pktmbuf) Name() string {
	return pm.name
}

func (pm *Pktmbuf) Mempool() *mempool.Mempool {
	return pm.m
}

func (pm *Pktmbuf) Free() {
	if pm.m != nil {
		pm.m.Free()
		pm.m = nil
	}

	// call given clean callback function if given during init
	if pm.clean != nil {
		pm.clean()
	}
}

// PktmbufStore represents a store of created DPDK Pktmbuf memory buffers
type PktmbufStore map[string]*Pktmbuf

func CreatePktmbufStore() PktmbufStore {
	return make(PktmbufStore)
}

func (ps PktmbufStore) Find(name string) *Pktmbuf {
	if name == "" {
		return nil
	}

	return ps[name]
}

// Create pktmbuf with corresponding pktmbuf mempool memory. Returns a pointer to a Pktmbuf
// structure or nil with error.
func (ps PktmbufStore) Create(
	name string,
	bufferSize uint,
	poolSize uint32,
	cacheSize uint32,
	cpuID int) (*Pktmbuf, error) {
	var pm Pktmbuf

	if ps.Find(name) != nil {
		return nil, errors.New("pktmbuf mempool with this name exists")
	}

	// create callback function called when record is freed
	clean := func() {
		delete(ps, name)
	}

	// initialize
	err := pm.Init(name, bufferSize, poolSize, cacheSize, cpuID, clean)
	if err != nil {
		return nil, err
	}

	// add node to list
	ps[name] = &pm

	return &pm, nil
}

// Delete given pktmbuf mempool from the store and free corresponding pktmbuf mempool memory
func (ps PktmbufStore) Delete(name string) {
	pm := ps.Find(name)

	if pm != nil {
		pm.Free()
	}
}

// Delete all pktmbuf mempools and free corresponding mempool memory
func (ps PktmbufStore) Clear() {
	for _, pm := range ps {
		pm.Free()
	}
}
