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
	"github.com/stolsma/go-p4pack/pkg/dpdkswx/pktmbuf"
)

type SourceParams struct {
	// Name of a valid PCAP file to read the input packets from.
	FileName string
	// Number of times to loop through the input PCAP file. Loop infinite times when set to 0.
	NLoops uint64
	// Maximum number of packets to read from the PCAP file. When 0, it is internally set to RTE_SWX_PORT_SOURCE_PKTS_MAX.
	// Once read from the PCAP file, the same packets are looped forever.
	NPktsMax uint32
	// Pktmbuf packet buffer pool. Must be valid.
	Pktmbuf *pktmbuf.Pktmbuf
}

// Source represents a Source device
type Source struct {
	*device.Device
	fileName string
	nLoops   uint64
	nPktsMax uint32
	pktmbuf  *pktmbuf.Pktmbuf
}

// Create and initialize Source device
func (s *Source) Init(name string, params *SourceParams, clean func()) error {
	// Node fill in
	s.Device = &device.Device{}
	s.SetType("SOURCE")
	s.SetName(name)
	s.fileName = params.FileName
	s.nLoops = params.NLoops
	s.nPktsMax = params.NPktsMax
	s.pktmbuf = params.Pktmbuf
	s.InitializeQueues(1, 0) // initialize queue setup for pipeline bind use
	s.SetClean(clean)

	return nil
}

// Free deletes the current Sink record and calls the clean callback function given at init
func (s *Source) Free() error {
	// call given clean callback function if given during init
	if s.Clean() != nil {
		s.Clean()()
	}

	return nil
}

type SwxPortSourceParams struct {
	rxParams  *C.struct_rte_swx_port_source_params
	paramsSet bool
	name      string
	mempool   *C.struct_rte_mempool
	fileName  string
	nLoops    uint64
	nPktsMax  uint32
}

func (e *SwxPortSourceParams) PortName() string {
	return e.name
}

func (e *SwxPortSourceParams) PortType() string {
	return "source"
}

func (e *SwxPortSourceParams) GetReaderParams() unsafe.Pointer {
	e.createCParams()
	return unsafe.Pointer(e.rxParams)
}

func (e *SwxPortSourceParams) GetWriterParams() unsafe.Pointer {
	return nil
}

func (e *SwxPortSourceParams) FreeParams() {
	e.freeCParams()
}

func (e *SwxPortSourceParams) createCParams() {
	if e.paramsSet {
		return
	}

	e.paramsSet = true
	e.rxParams = &C.struct_rte_swx_port_source_params{
		pool:       e.mempool,
		file_name:  C.CString(e.fileName),
		n_loops:    C.ulong(e.nLoops),
		n_pkts_max: C.uint(e.nPktsMax),
	}
}

func (e *SwxPortSourceParams) freeCParams() {
	if !e.paramsSet {
		return
	}

	e.paramsSet = false
	C.free(unsafe.Pointer(e.rxParams.file_name))
	e.rxParams = nil
}

// bind to given pipeline input port
func (s *Source) BindToPipelineInputPort(pl *pipeline.Pipeline, portID int, rxq uint16, bsz uint) error {
	if _, plp, err := s.GetRxQueue(rxq); err != nil {
		return err
	} else if plp != device.NotBound {
		return errors.New("port already bound")
	}

	params := &SwxPortSourceParams{
		name:     s.Name(),
		fileName: s.fileName,
		mempool:  (*C.struct_rte_mempool)(unsafe.Pointer(s.pktmbuf.Mempool())),
		nLoops:   s.nLoops,
		nPktsMax: s.nPktsMax,
	}
	if err := pl.PortInConfig(portID, params); err != nil {
		return err
	}

	return s.SetRxQueue(rxq, pl.GetName(), portID)
}

// bind to given pipeline output port
func (s *Source) BindToPipelineOutputPort(pl *pipeline.Pipeline, portID int, txq uint16, bsz uint) error {
	return errors.New("can't bind source device to pipeline output port")
}
