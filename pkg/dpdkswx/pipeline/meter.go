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
#include <rte_meter.h>
*/
import "C"
import (
	"errors"
	"fmt"
	"unsafe"

	"github.com/stolsma/go-p4pack/pkg/dpdkswx/common"
)

type MeterProfile struct {
	name string
	cir  uint64
	pir  uint64
	cbs  uint64
	pbs  uint64
}

func (mp *MeterProfile) GetName() string {
	return mp.name
}

func CreateMeterProfile(name string, cir uint64, pir uint64, cbs uint64, pbs uint64) *MeterProfile {
	var mp = &MeterProfile{
		name: name,
		cir:  cir,
		pir:  pir,
		cbs:  cbs,
		pbs:  pbs,
	}
	return mp
}

// MeterProfileStore represents a store of MeterProfile records
type MeterProfileStore struct {
	pl    *Pipeline
	store map[string]*MeterProfile
}

func CreateMeterProfileStore(pl *Pipeline) *MeterProfileStore {
	return &MeterProfileStore{
		pl:    pl,
		store: make(map[string]*MeterProfile),
	}
}

func (mps *MeterProfileStore) FindName(name string) *MeterProfile {
	if name == "" {
		return nil
	}

	return mps.store[name]
}

// Meter profile add
//
// Returns nil on success or the following error codes otherwise:
//
//	-EINVAL = Invalid argument
//	-ENOMEM = Not enough space/cannot allocate memory
//	-EEXIST = Meter profile with this name already exists
func (mps *MeterProfileStore) Add(mp *MeterProfile) error {
	var params = C.struct_rte_meter_trtcm_params{
		cir: C.ulong(mp.cir),
		pir: C.ulong(mp.pir),
		cbs: C.ulong(mp.cbs),
		pbs: C.ulong(mp.pbs),
	}
	cMPN := C.CString(mp.GetName())
	defer C.free(unsafe.Pointer(cMPN))

	if result := C.rte_swx_ctl_meter_profile_add(mps.pl.p, cMPN, &params); result != 0 {
		return common.Err(result)
	}

	mps.store[mp.GetName()] = mp

	return nil
}

// Meter profile delete
//
// Returns nil on success or the following error codes otherwise:
//
//	-EINVAL = Invalid argument
//	-EBUSY = Meter profile is currently in use
func (mps *MeterProfileStore) Delete(mpName string) error {
	cMPN := C.CString(mpName)
	defer C.free(unsafe.Pointer(cMPN))

	if result := C.rte_swx_ctl_meter_profile_delete(mps.pl.p, cMPN); result != 0 {
		return common.Err(result)
	}

	delete(mps.store, mpName)

	return nil
}

// Iterate through all stored meter profiles. If the callback function returns error the iteration stops and the error
// will be returned. If all profiles are itereated nil will be returned.
func (mps *MeterProfileStore) Iterate(fn func(key string, mp *MeterProfile) error) error {
	for k, v := range mps.store {
		if err := fn(k, v); err != nil {
			return err
		}
	}
	return nil
}

// Delete all MeterProfile records
func (mps *MeterProfileStore) Clear() error {
	for _, mp := range mps.store {
		if err := mps.Delete(mp.GetName()); err != nil {
			return err
		}
	}

	return nil
}

type Meter struct {
	pipeline *Pipeline // parent pipeline
	index    uint      // Index in swx_pipeline meter store
	name     string    // Meter name.
	size     int       // Meter size parameter.
}

// Initialize meter record from pipeline
func (m *Meter) Init(p *Pipeline, index uint) error {
	meterInfo, err := p.MetarrayInfoGet(index)
	if err != nil {
		return err
	}

	// initalize generic table attributes
	m.pipeline = p
	m.index = index
	m.name = meterInfo.GetName()
	m.size = meterInfo.GetSize()

	return nil
}

func (m *Meter) Clear() {
	// TODO check if all memory related to this structure is freed
}

func (m *Meter) GetIndex() uint {
	return m.index
}

func (m *Meter) GetName() string {
	return m.name
}

func (m *Meter) GetSize() int {
	return m.size
}

// Reset meter
//
// Reset a meter within a given meter array (index) to use the default profile that causes all the input packets to be
// colored as green. It is the responsibility of the control plane to make sure this meter is not used by the data plane
// pipeline before calling this function. Returns nil on success or the following error codes otherwise:
//
//	-EINVAL = Invalid argument
func (m *Meter) Reset(index uint32) error {
	cMeter := C.CString(m.name)
	defer C.free(unsafe.Pointer(cMeter))

	if result := C.rte_swx_ctl_meter_reset(m.pipeline.p, cMeter, C.uint(index)); result != 0 {
		return common.Err(result)
	}

	return nil
}

const (
	ColorGreen  = C.RTE_COLOR_GREEN  // Green
	ColorYellow = C.RTE_COLOR_YELLOW // Yellow
	ColorRed    = C.RTE_COLOR_RED    // Red
	Colors      = C.RTE_COLORS       // Number of colors
)

// Meter set
//
// Set a meter within a given meter array to use a specific profile. It is the responsibility of the control plane to
// make sure this meter is not used by the data plane pipeline before calling this function. Returns nil on success or
// the following error codes otherwise:
//
//	-EINVAL = Invalid argument
func (m *Meter) Set(index uint32, profile string) error {
	cMeter := C.CString(m.name)
	defer C.free(unsafe.Pointer(cMeter))
	cProfile := C.CString(profile)
	defer C.free(unsafe.Pointer(cProfile))

	if result := C.rte_swx_ctl_meter_set(m.pipeline.p, cMeter, C.uint(index), cProfile); result != 0 {
		return common.Err(result)
	}

	return nil
}

// Meter statistics counters.
type MeterStats C.struct_rte_swx_ctl_meter_stats

func (ms *MeterStats) Pkts(color uint) (uint64, error) {
	if color >= Colors {
		return 0, errors.New("color index to large")
	}

	return uint64(ms.n_pkts[color]), nil
}

func (ms *MeterStats) Bytes(color uint) (uint64, error) {
	if color >= Colors {
		return 0, errors.New("color index to large")
	}

	return uint64(ms.n_bytes[color]), nil
}

// Meter statistics counters read
//
// Returns nil on success or the following error codes otherwise:
//
//	-EINVAL = Invalid argument
func (m *Meter) Read(index uint32, profile string) (*MeterStats, error) {
	var stats = &MeterStats{}
	cMeter := C.CString(m.name)
	defer C.free(unsafe.Pointer(cMeter))

	if result := C.rte_swx_ctl_meter_stats_read(m.pipeline.p, cMeter, C.uint(index),
		(*C.struct_rte_swx_ctl_meter_stats)(stats),
	); result != 0 {
		return nil, common.Err(result)
	}

	return stats, nil
}

// MeterStore represents a store of Meter records
type MeterStore map[string]*Meter

func CreateMeterStore() MeterStore {
	return make(MeterStore)
}

func (ms MeterStore) FindName(name string) *Meter {
	if name == "" {
		return nil
	}

	return ms[name]
}

func (ms MeterStore) CreateFromPipeline(p *Pipeline) error {
	pipelineInfo, err := p.PipelineInfoGet()
	if err != nil {
		return err
	}

	for i := uint(0); i < pipelineInfo.GetNMetarrays(); i++ {
		var meter Meter

		err := meter.Init(p, i)
		if err != nil {
			return fmt.Errorf("Meterstore.CreateFromPipeline error: %d", err)
		}
		ms.Add(&meter)
	}

	return nil
}

func (ms MeterStore) Add(meter *Meter) {
	ms[meter.GetName()] = meter
}

func (ms MeterStore) ForEach(fn func(key string, meter *Meter) error) error {
	for k, v := range ms {
		if err := fn(k, v); err != nil {
			return err
		}
	}
	return nil
}

// Delete all Meter records and free corresponding memory if required
func (ms MeterStore) Clear() {
	for _, meter := range ms {
		meter.Clear()
		delete(ms, meter.GetName())
	}
}
