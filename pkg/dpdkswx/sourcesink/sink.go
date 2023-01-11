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
	s.InitializeQueues(0, 1) // initialize queue setup for pipeline bind use
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

type SwxPortSinkParams struct {
	txParams  *C.struct_rte_swx_port_sink_params
	paramsSet bool
	name      string
	fileName  string
}

func (e *SwxPortSinkParams) PortName() string {
	return e.name
}

func (e *SwxPortSinkParams) PortType() string {
	return "sink"
}

func (e *SwxPortSinkParams) GetReaderParams() unsafe.Pointer {
	return nil
}

func (e *SwxPortSinkParams) GetWriterParams() unsafe.Pointer {
	e.createCParams()
	return unsafe.Pointer(e.txParams)
}

func (e *SwxPortSinkParams) FreeParams() {
	e.freeCParams()
}

func (e *SwxPortSinkParams) createCParams() {
	if e.paramsSet {
		return
	}

	e.paramsSet = true
	e.txParams = &C.struct_rte_swx_port_sink_params{
		file_name: C.CString(e.fileName),
	}
}

func (e *SwxPortSinkParams) freeCParams() {
	if !e.paramsSet {
		return
	}

	e.paramsSet = false
	C.free(unsafe.Pointer(e.txParams.file_name))
	e.txParams = nil
}

// bind to given pipeline input port
func (s *Sink) BindToPipelineInputPort(pl *pipeline.Pipeline, portID int, rxq uint16, bsz uint) error {
	return errors.New("can't bind sink device to pipeline input port")
}

// bind to given pipeline output port
func (s *Sink) BindToPipelineOutputPort(pl *pipeline.Pipeline, portID int, txq uint16, bsz uint) error {
	if _, plp, err := s.GetTxQueue(txq); err != nil {
		return err
	} else if plp != device.NotBound {
		return errors.New("port already bound")
	}

	params := &SwxPortSinkParams{
		name:     s.Name(),
		fileName: s.fileName,
	}
	if err := pl.PortOutConfig(portID, params); err != nil {
		return err
	}

	return s.SetTxQueue(txq, pl.GetName(), portID)
}
