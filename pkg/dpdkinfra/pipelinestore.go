// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package dpdkinfra

import "github.com/stolsma/go-p4dpdk-vswitch/pkg/dpdkswx"

// PipelineStore represents a store of created Pipeline records
type PipelineStore map[string]*dpdkswx.Pipeline

func CreatePipelineStore() PipelineStore {
	return make(PipelineStore)
}

// Find a Pipeline in this store
func (pls PipelineStore) Find(name string) *dpdkswx.Pipeline {
	if name == "" {
		return nil
	}

	return pls[name]
}

// Create Pipeline. Returns a pointer to a Pipeline structure or nil with error.
func (pls PipelineStore) Create(name string, numaNode int) (*dpdkswx.Pipeline, error) {
	var pipeline dpdkswx.Pipeline

	// create callback function called when record is freed
	clean := func() {
		delete(pls, name)
	}

	// initialize pipeline record
	pipeline.Init(name, numaNode, clean)

	// add node to list
	pls[name] = &pipeline

	return &pipeline, nil
}

// Delete all Pipeline records and free corresponding memory
func (pls PipelineStore) Clear() {
	for _, pipeline := range pls {
		pipeline.Free()
	}
}
