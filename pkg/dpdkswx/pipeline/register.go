// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package pipeline

/*
#include <stdlib.h>
#include <string.h>
#include <netinet/in.h>
#include <sys/ioctl.h>
#include <fcntl.h>
#include <unistd.h>
#include <stdint.h>
#include <sys/queue.h>

#include <rte_swx_pipeline.h>
#include <rte_swx_ctl.h>
*/
import "C"
import (
	"fmt"
	"unsafe"

	"github.com/stolsma/go-p4pack/pkg/dpdkswx/common"
)

type Register struct {
	pipeline *Pipeline // parent pipeline
	index    uint      // Index in swx_pipeline register store
	name     string    // Register name.
	size     int       // Register size parameter.
}

// Initialize register record from pipeline
func (r *Register) Init(p *Pipeline, index uint) error {
	regInfo, err := p.RegarrayInfoGet(index)
	if err != nil {
		return err
	}

	// initalize generic table attributes
	r.pipeline = p
	r.index = index
	r.name = regInfo.GetName()
	r.size = regInfo.GetSize()

	return nil
}

func (r *Register) Clear() {
	// TODO check if all memory related to this structure is freed
}

func (r *Register) GetIndex() uint {
	return r.index
}

func (r *Register) GetName() string {
	return r.name
}

func (r *Register) GetSize() int {
	return r.size
}

// Register read
//
// Read the current register index (index) value. Returns nil on success or the following error codes otherwise:
//
//	-EINVAL = Invalid argument
func (r *Register) RegisterRead(index uint32) (uint64, error) {
	var value C.ulong
	cRegister := C.CString(r.name)
	defer C.free(unsafe.Pointer(cRegister))

	if result := C.rte_swx_ctl_pipeline_regarray_read(r.pipeline.p, cRegister, C.uint(index), &value); result != 0 {
		return 0, common.Err(result)
	}

	return uint64(value), nil
}

// Register write
//
// Write the current register index (index) value. Returns nil on success or the following error codes otherwise:
//
//	-EINVAL = Invalid argument
func (r *Register) RegisterWrite(index uint32, value uint64) error {
	cRegister := C.CString(r.name)
	defer C.free(unsafe.Pointer(cRegister))

	if result := C.rte_swx_ctl_pipeline_regarray_write(r.pipeline.p, cRegister, C.uint(index), C.ulong(value)); result != 0 {
		return common.Err(result)
	}

	return nil
}

// RegisterStore represents a store of Register records
type RegisterStore map[string]*Register

func CreateRegisterStore() RegisterStore {
	return make(RegisterStore)
}

func (rs RegisterStore) FindName(name string) *Register {
	if name == "" {
		return nil
	}

	return rs[name]
}

func (rs RegisterStore) CreateFromPipeline(p *Pipeline) error {
	pipelineInfo, err := p.PipelineInfoGet()
	if err != nil {
		return err
	}

	for i := uint(0); i < pipelineInfo.GetNRegarrays(); i++ {
		var register Register

		err := register.Init(p, i)
		if err != nil {
			return fmt.Errorf("Registerstore.CreateFromPipeline error: %d", err)
		}
		rs.Add(&register)
	}

	return nil
}

func (rs RegisterStore) Add(register *Register) {
	rs[register.GetName()] = register
}

func (rs RegisterStore) ForEach(fn func(key string, register *Register) error) error {
	for k, v := range rs {
		if err := fn(k, v); err != nil {
			return err
		}
	}
	return nil
}

// Delete all Register records and free corresponding memory if required
func (rs RegisterStore) Clear() {
	for _, register := range rs {
		register.Clear()
		delete(rs, register.GetName())
	}
}
