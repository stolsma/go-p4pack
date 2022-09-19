// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package dpdkinfra

import "github.com/stolsma/go-p4dpdk-vswitch/pkg/dpdkswx"

// TableStore represents a store of Table records
type TableStore map[string]*dpdkswx.Table

func CreateTableStore() TableStore {
	return make(TableStore)
}

func (ts TableStore) Find(name string) *dpdkswx.Table {
	if name == "" {
		return nil
	}

	return ts[name]
}

// Create Table record. Returns a pointer to a Table structure or nil with error.
func (ts TableStore) Create(p *dpdkswx.Pipeline, tableId uint) (*dpdkswx.Table, error) {
	var table dpdkswx.Table

	// create callback function called when record is freed
	clean := func() {
		//		delete(ts, name)
	}

	err := table.Init(p, tableId, clean)
	return &table, err
}

// Delete all Table records and free corresponding memory if required
func (ts TableStore) Clear() {
	for _, table := range ts {
		table.Clear()
	}
}
