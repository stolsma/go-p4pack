// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package dpdkinfra

import (
	"errors"

	"github.com/stolsma/go-p4pack/pkg/dpdkswx/ethdev"
)

// EthdevStore represents a store of created DPDK Ethdev interfaces
type EthdevStore map[string]*ethdev.Ethdev

func CreateEthdevStore() EthdevStore {
	return make(EthdevStore)
}

func (eds EthdevStore) Find(name string) *ethdev.Ethdev {
	if name == "" {
		return nil
	}

	return eds[name]
}

// Create Ethdev. Returns a pointer to a Ethdev structure or nil with error.
func (eds EthdevStore) Create(name string, params *ethdev.LinkParams) (*ethdev.Ethdev, error) {
	var ethdev ethdev.Ethdev

	if eds.Find(name) != nil {
		return nil, errors.New("ethdev with this name exists")
	}

	// create callback function called when record is freed
	clean := func() {
		delete(eds, name)
	}

	// initialize
	err := ethdev.Init(name, params, clean)
	if err != nil {
		return nil, err
	}

	// add node to list
	eds[name] = &ethdev

	return &ethdev, nil
}

// Delete given Ethdev from the store and free corresponding Ethdev memory
func (eds EthdevStore) Delete(name string) {
	ethdev := eds.Find(name)

	if ethdev != nil {
		ethdev.Free()
	}
}

// Delete all Ethdev and free corresponding memory
func (eds EthdevStore) Clear() {
	for _, ethdev := range eds {
		ethdev.Free()
	}
}

func (eds EthdevStore) Iterate(fn func(key string, ethdev *ethdev.Ethdev) error) error {
	if fn != nil {
		for k, v := range eds {
			if err := fn(k, v); err != nil {
				return err
			}
		}
	} else {
		return errors.New("no function to call")
	}
	return nil
}
