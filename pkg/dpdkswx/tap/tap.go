// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package tap

/*
#include <stdlib.h>
#include <string.h>
#include <bsd/string.h>
#include <netinet/in.h>
#include <linux/if.h>
#include <linux/if_tun.h>
#include <sys/ioctl.h>
#include <fcntl.h>
#include <unistd.h>

#include <rte_swx_port_fd.h>

#define TAP_DEV		"/dev/net/tun"

int tap_create(const char *name) {
	struct ifreq ifr;
	int fd, status;

	// Resource create
	fd = open(TAP_DEV, O_RDWR | O_NONBLOCK);
	if (fd < 0)
		return fd;

	memset(&ifr, 0, sizeof(ifr));
	ifr.ifr_flags = IFF_TAP | IFF_NO_PI; // No packet information
	strlcpy(ifr.ifr_name, name, IFNAMSIZ);

	status = ioctl(fd, TUNSETIFF, (void *) &ifr);
	if (status < 0) {
		close(fd);
		return status;
	}

  return fd;
}

*/
import "C"

import (
	"errors"
	"unsafe"

	"github.com/stolsma/go-p4pack/pkg/dpdkswx/device"
	"github.com/stolsma/go-p4pack/pkg/dpdkswx/pipeline"
	"github.com/stolsma/go-p4pack/pkg/dpdkswx/pktmbuf"
)

type Params struct {
	Mtu     int
	Pktmbuf *pktmbuf.Pktmbuf
}

// Tap represents a Tap record stored in a tap store
type Tap struct {
	*device.Device
	fd      C.int
	mtu     int
	pktmbuf *pktmbuf.Pktmbuf
}

// Create Tap interface. Returns a pointer to a Tap structure or nil with error.
func (tap *Tap) Init(name string, params *Params, clean func()) error {
	// create fd of tap interface
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	fd, err := C.tap_create(cname)
	if err != nil {
		return err
	}

	// Node fill in
	tap.Device = &device.Device{}
	tap.SetType("TAP")
	tap.SetName(name)
	tap.fd = fd
	tap.mtu = params.Mtu
	tap.pktmbuf = params.Pktmbuf
	tap.InitializeQueues(1, 1) // initialize queue setup for pipeline bind use
	tap.SetClean(clean)

	return nil
}

func (tap *Tap) Type() string {
	return "TAP"
}

// Fd returns the File descripter of the Tap interface
func (tap *Tap) Fd() C.int {
	return tap.fd
}

// Free deletes the current Tap record and calls the clean callback function given at init
func (tap *Tap) Free() error {
	// TODO remove TAP interface from the system
	// call given clean callback function if given during init
	if tap.Clean() != nil {
		tap.Clean()()
	}

	return nil
}

type SwxPortTapParams struct {
	rxParams  *C.struct_rte_swx_port_fd_reader_params
	txParams  *C.struct_rte_swx_port_fd_writer_params
	paramsSet bool
	name      string

	fd      C.int
	mtu     uint
	mempool *C.struct_rte_mempool
	bsz     uint
}

func (e *SwxPortTapParams) PortName() string {
	return e.name
}

func (e *SwxPortTapParams) PortType() string {
	return "fd"
}

func (e *SwxPortTapParams) GetReaderParams() unsafe.Pointer {
	e.createCParams()
	return unsafe.Pointer(e.rxParams)
}

func (e *SwxPortTapParams) GetWriterParams() unsafe.Pointer {
	e.createCParams()
	return unsafe.Pointer(e.txParams)
}

func (e *SwxPortTapParams) FreeParams() {
	e.freeCParams()
}

func (e *SwxPortTapParams) createCParams() {
	if e.paramsSet {
		return
	}

	e.paramsSet = true
	e.rxParams = &C.struct_rte_swx_port_fd_reader_params{
		fd:         e.fd,
		mtu:        (C.uint)(e.mtu),
		mempool:    e.mempool,
		burst_size: (C.uint)(e.bsz),
	}
	e.txParams = &C.struct_rte_swx_port_fd_writer_params{
		fd:         e.fd,
		burst_size: (C.uint)(e.bsz),
	}
}

func (e *SwxPortTapParams) freeCParams() {
	if !e.paramsSet {
		return
	}

	e.paramsSet = false
	e.rxParams = nil
	e.txParams = nil
}

// bind to given pipeline input port. A tap has 1 queue so only queue number 0 is valid.
func (tap *Tap) BindToPipelineInputPort(pl *pipeline.Pipeline, portID int, rxq uint16, bsz uint) error {
	if _, plp, err := tap.GetRxQueue(rxq); err != nil {
		return err
	} else if plp != device.NotBound {
		return errors.New("port already bound")
	}

	params := &SwxPortTapParams{
		name:    tap.Name(),
		fd:      tap.Fd(),
		mempool: (*C.struct_rte_mempool)(unsafe.Pointer(tap.pktmbuf.Mempool())),
		mtu:     (uint)(tap.mtu),
		bsz:     bsz,
	}
	if err := pl.PortInConfig(portID, params); err != nil {
		return err
	}

	return tap.SetRxQueue(rxq, pl.GetName(), portID)
}

// bind to given pipeline output port. A tap has 1 queue so only queue number 0 is valid.
func (tap *Tap) BindToPipelineOutputPort(pl *pipeline.Pipeline, portID int, txq uint16, bsz uint) error {
	if _, plp, err := tap.GetTxQueue(txq); err != nil {
		return err
	} else if plp != device.NotBound {
		return errors.New("port already bound")
	}

	params := &SwxPortTapParams{
		name: tap.Name(),
		fd:   tap.Fd(),
		bsz:  bsz,
	}
	if err := pl.PortOutConfig(portID, params); err != nil {
		return err
	}

	return tap.SetTxQueue(txq, pl.GetName(), portID)
}
