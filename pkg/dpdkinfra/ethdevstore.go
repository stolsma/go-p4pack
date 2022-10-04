// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package dpdkinfra

import (
	"errors"

	"github.com/stolsma/go-p4pack/pkg/dpdkswx"
)

// EthdevStore represents a store of created DPDK Ethdev interfaces
type EthdevStore map[string]*dpdkswx.Ethdev

func CreateEthdevStore() EthdevStore {
	return make(EthdevStore)
}

func (ps EthdevStore) Find(name string) *dpdkswx.Ethdev {
	if name == "" {
		return nil
	}

	return ps[name]
}

// Create Ethdev. Returns a pointer to a Ethdev structure or nil with error.
func (ps EthdevStore) Create(name string, params *dpdkswx.EthdevParams) (*dpdkswx.Ethdev, error) {

	var ethdev dpdkswx.Ethdev

	if ps.Find(name) != nil {
		return nil, errors.New("ethdev with this name exists")
	}

	// create callback function called when record is freed
	clean := func() {
		delete(ps, name)
	}

	// initialize
	err := ethdev.Init(name, params, clean)
	if err != nil {
		return nil, err
	}

	// add node to list
	ps[name] = &ethdev

	return &ethdev, nil
}

// Delete given Ethdev from the store and free corresponding Ethdev memory
func (ps EthdevStore) Delete(name string) {
	ethdev := ps.Find(name)

	if ethdev != nil {
		ethdev.Free()
	}
}

// Delete all Ethdev and free corresponding memory
func (ps EthdevStore) Clear() {
	for _, ethdev := range ps {
		ethdev.Free()
	}
}
