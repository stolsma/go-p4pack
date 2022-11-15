// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package dpdkinfra

import (
	"errors"

	"github.com/stolsma/go-p4pack/pkg/dpdkswx/pktmbuf"
)

// PktmbufStore represents a store of created DPDK Pktmbuf memory buffers
type PktmbufStore map[string]*pktmbuf.Pktmbuf

func CreatePktmbufStore() PktmbufStore {
	return make(PktmbufStore)
}

func (ps PktmbufStore) Find(name string) *pktmbuf.Pktmbuf {
	if name == "" {
		return nil
	}

	return ps[name]
}

// Create pktmbuf with corresponding pktmbuf mempool memory. Returns a pointer to a Pktmbuf
// structure or nil with error.
func (ps PktmbufStore) Create(
	name string,
	bufferSize uint32,
	poolSize uint32,
	cacheSize uint32,
	cpuID int) (*pktmbuf.Pktmbuf, error) {
	var pm pktmbuf.Pktmbuf

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
