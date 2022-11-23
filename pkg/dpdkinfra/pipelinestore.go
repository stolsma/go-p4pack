// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package dpdkinfra

import (
	"errors"

	"github.com/stolsma/go-p4pack/pkg/dpdkswx/pipeline"
)

// PipelineStore represents a store of created Pipeline records
type PipelineStore map[string]*pipeline.Pipeline

func CreatePipelineStore() PipelineStore {
	return make(PipelineStore)
}

// Find a Pipeline in this store
func (pls PipelineStore) Find(name string) *pipeline.Pipeline {
	if name == "" {
		return nil
	}

	return pls[name]
}

// Create Pipeline. Returns a pointer to a Pipeline structure or nil with error.
func (pls PipelineStore) Create(name string, numaNode int) (*pipeline.Pipeline, error) {
	var pl pipeline.Pipeline

	// create callback function called when record is freed
	clean := func() {
		log.Infof("Remove pipeline %s from store", name)
		delete(pls, name)
	}

	// initialize pipeline record
	pl.Init(name, numaNode, clean)

	// add node to list
	pls[name] = &pl

	return &pl, nil
}

// Delete all Pipeline records and free corresponding memory
func (pls PipelineStore) Clear() {
	for _, pl := range pls {
		pl.Free()
	}
}

func (pls PipelineStore) Iterate(fn func(key string, pipeline *pipeline.Pipeline) error) error {
	if fn != nil {
		for k, v := range pls {
			if err := fn(k, v); err != nil {
				return err
			}
		}
	} else {
		return errors.New("no function to call")
	}
	return nil
}
