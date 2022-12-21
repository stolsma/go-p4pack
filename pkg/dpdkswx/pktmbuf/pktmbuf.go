// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package pktmbuf

/*
#include <rte_mempool.h>
#include <rte_mbuf.h>

*/
import "C"
import (
	"errors"

	"github.com/stolsma/go-p4pack/pkg/dpdkswx"
	"github.com/stolsma/go-p4pack/pkg/dpdkswx/swxruntime"
	"github.com/stolsma/go-p4pack/pkg/logging"
	"github.com/yerden/go-dpdk/mempool"
)

var log logging.Logger

func init() {
	// keep the logger up to date, also after new log config
	logging.Register("dpdkswx/pktmbuf", func(logger logging.Logger) {
		log = logger
	})
}

const RteMbufDefaultBufSize uint = C.RTE_MBUF_DEFAULT_BUF_SIZE
const RteMbufDefaultDataroom uint = C.RTE_MBUF_DEFAULT_DATAROOM
const RtePktmbufHeadroom uint = C.RTE_PKTMBUF_HEADROOM
const SizeofRteMbuf uint = C.sizeof_struct_rte_mbuf

const BufferSizeMin uint = SizeofRteMbuf + RtePktmbufHeadroom

// Pktmbuf represents a DPDK Packet memory buffer (pktmbuf) initialized memory pool (mempool)
type Pktmbuf struct {
	name       string
	m          *mempool.Mempool
	bufferSize uint
	clean      func()
}

// Create pktmbuf with corresponding pktmbuf mempool memory. Returns a pointer to a Pktmbuf
// structure or nil with error.
func (pm *Pktmbuf) Init(
	name string,
	bufferSize uint,
	poolSize uint32,
	cacheSize uint32,
	cpuSocketID int,
	clean func(),
) error {
	// Check input params
	if name == "" {
		return errors.New("no name given")
	}

	if bufferSize < BufferSizeMin {
		return errors.New("buffer size to small")
	}

	if poolSize == 0 {
		return errors.New("pool size is 0")
	}

	// create PktMbufpool on main DPDK Lcore to prevent problems
	var md *mempool.Mempool
	err := dpdkswx.Runtime.ExecOnMain(func(*swxruntime.MainCtx) (err error) {
		md, err = mempool.CreateMbufPool(
			name,
			poolSize,
			uint16(bufferSize),
			mempool.OptSocket(cpuSocketID),
			mempool.OptCacheSize(cacheSize),
			mempool.OptPrivateDataSize(0), // for each Mbuf
		)
		return err
	})
	if err != nil {
		return err
	}

	// Node fill in
	pm.name = name
	pm.m = md
	pm.bufferSize = bufferSize
	pm.clean = clean

	return nil
}

func (pm *Pktmbuf) Name() string {
	return pm.name
}

func (pm *Pktmbuf) Mempool() *mempool.Mempool {
	return pm.m
}

func (pm *Pktmbuf) Free() error {
	if pm.m != nil {
		pm.m.Free()
		pm.m = nil
	}

	// call given clean callback function if given during init
	if pm.clean != nil {
		pm.clean()
	}

	return nil
}
