// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package sourcesink

/*
#include <stdlib.h>

#include <rte_swx_port_source_sink.h>

*/
import "C"
import (
	"errors"
	"unsafe"

	"github.com/stolsma/go-p4pack/pkg/dpdkswx/device"
	"github.com/stolsma/go-p4pack/pkg/dpdkswx/pipeline"
)

type SinkParams struct {
	// Name of a valid PCAP file to write the output packets to. When NULL, all the output packets are dropped instead
	// of being saved to a PCAP file.
	FileName string
}

// Sink represents a Sink device
type Sink struct {
	*device.Device
	fileName string
}

// Create and initialize Sink device
func (s *Sink) Init(name string, params *SinkParams, clean func()) error {
	// Node fill in
	s.Device = &device.Device{}
	s.SetType("SINK")
	s.SetName(name)
	s.fileName = params.FileName

	s.SetPipelineOutPort(device.NotBound)
	s.SetPipelineInPort(device.NotBound)
	s.SetClean(clean)

	return nil
}

// Free deletes the current Sink record and calls the clean callback function given at init
func (s *Sink) Free() error {
	// call given clean callback function if given during init
	if s.Clean() != nil {
		s.Clean()()
	}

	return nil
}

// bind to given pipeline input port
func (s *Sink) BindToPipelineInputPort(pl *pipeline.Pipeline, portID int, rxq uint, bsz uint) error {
	return errors.New("can't bind sink device to pipeline input port")
}

// bind to given pipeline output port
func (s *Sink) BindToPipelineOutputPort(pl *pipeline.Pipeline, portID int, txq uint, bsz uint) error {
	var params C.struct_rte_swx_port_sink_params

	if s.PipelineOutPort() != device.NotBound {
		return errors.New("port already bound")
	}
	s.SetPipelineOut(pl.GetName())
	s.SetPipelineOutPort(portID)

	params.file_name = C.CString(s.fileName)
	defer C.free(unsafe.Pointer(params.file_name))

	return pl.PortOutConfig(portID, "sink", unsafe.Pointer(&params))
}
