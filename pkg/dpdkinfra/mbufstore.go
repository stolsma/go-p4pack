// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package dpdkinfra

import (
	"errors"

	"github.com/stolsma/go-p4pack/pkg/dpdkswx"
)

// PktmbufStore represents a store of created DPDK Pktmbuf memory buffers
type PktmbufStore map[string]*dpdkswx.Pktmbuf

func CreatePktmbufStore() PktmbufStore {
	return make(PktmbufStore)
}

func (ps PktmbufStore) Find(name string) *dpdkswx.Pktmbuf {
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
	cpuID int) (*dpdkswx.Pktmbuf, error) {
	var pktmbuf dpdkswx.Pktmbuf

	if ps.Find(name) != nil {
		return nil, errors.New("pktmbuf mempool with this name exists")
	}

	// create callback function called when record is freed
	clean := func() {
		delete(ps, name)
	}

	// initialize
	err := pktmbuf.Init(name, bufferSize, poolSize, cacheSize, cpuID, clean)
	if err != nil {
		return nil, err
	}

	// add node to list
	ps[name] = &pktmbuf

	return &pktmbuf, nil
}

// Delete given pktmbuf mempool from the store and free corresponding pktmbuf mempool memory
func (ps PktmbufStore) Delete(name string) {
	pktmbuf := ps.Find(name)

	if pktmbuf != nil {
		pktmbuf.Free()
	}
}

// Delete all pktmbuf mempools and free corresponding mempool memory
func (ps PktmbufStore) Clear() {
	for _, pktbuf := range ps {
		pktbuf.Free()
	}
}
