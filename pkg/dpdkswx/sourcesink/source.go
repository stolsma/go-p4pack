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

	s.SetPipelineOutPort(device.NotBound)
	s.SetPipelineInPort(device.NotBound)
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

// bind to given pipeline input port
func (s *Source) BindToPipelineInputPort(pl *pipeline.Pipeline, portID int, rxq uint, bsz uint) error {
	var params C.struct_rte_swx_port_source_params

	if s.PipelineInPort() != device.NotBound {
		return errors.New("port already bound")
	}
	s.SetPipelineIn(pl.GetName())
	s.SetPipelineInPort(portID)

	params.file_name = C.CString(s.fileName)
	defer C.free(unsafe.Pointer(params.file_name))
	params.n_loops = C.ulong(s.nLoops)
	params.n_pkts_max = C.uint(s.nPktsMax)
	params.pool = (*C.struct_rte_mempool)(unsafe.Pointer(s.pktmbuf.Mempool()))

	return pl.PortInConfig(portID, "source", unsafe.Pointer(&params))
}

// bind to given pipeline output port
func (s *Source) BindToPipelineOutputPort(pl *pipeline.Pipeline, portID int, txq uint, bsz uint) error {
	return errors.New("can't bind source device to pipeline output port")
}
