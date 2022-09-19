// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package dpdkswx

/*
#cgo pkg-config: libdpdk

#include <stdlib.h>
#include <stdint.h>
#include <string.h>

#include <rte_ethdev.h>

*/
import "C"
import (
	"errors"
	"fmt"
	"unsafe"

	lled "github.com/stolsma/go-p4dpdk-vswitch/pkg/dpdkswx/ethdev"
)

const RETA_CONF_SIZE = (C.RTE_ETH_RSS_RETA_SIZE_512 / C.RTE_ETH_RETA_GROUP_SIZE)

const LINK_RXQ_RSS_MAX = 16

type EthdevParamsRss struct {
	queue_id [LINK_RXQ_RSS_MAX]uint32
	n_queues uint32
}

type EthdevParams struct {
	devName           string
	devArgs           string
	portId            uint16 // Valid only when *dev_name* is NULL.
	devHotplugEnabled bool

	rx struct {
		mtu        uint32
		n_queues   uint16
		queue_size uint32
		mempool    *Pktmbuf
		rss        *EthdevParamsRss
	}

	tx struct {
		n_queues   uint16
		queue_size uint32
	}

	promiscuous bool
}

// Ethdev represents a Ethdev record
type Ethdev struct {
	name     string
	dev_name string
	portId   lled.Port
	nRxQ     uint16
	nTxQ     uint16
	clean    func()
}

// Create Link interface. Returns a pointer to a Link structure or nil with error.
func (ethdev *Ethdev) Init(name string, params *EthdevParams, clean func()) error {
	var port_info lled.DevInfo
	var port_id lled.Port
	var rss *EthdevParamsRss //*C.struct_link_params_rss
	var mempool *Pktmbuf
	var status C.int
	var res error

	// TODO add all params to check!
	// Check input params
	if (name == "") || (params == nil) || (params.rx.n_queues == 0) ||
		(params.rx.queue_size == 0) || (params.tx.n_queues == 0) || (params.tx.queue_size == 0) {
		return nil
	}

	//	cname := C.CString(name)
	//	defer C.free(unsafe.Pointer(cname))
	devName := C.CString(params.devName)
	defer C.free(unsafe.Pointer(devName))
	devArgs := C.CString(params.devArgs)
	defer C.free(unsafe.Pointer(devArgs))

	// Performing Device Hotplug and valid for only VDEVs
	if params.devHotplugEnabled {
		vdev := C.CString("vdev")
		status = C.rte_eal_hotplug_add(vdev, devName, devArgs)
		C.free(unsafe.Pointer(vdev))
		if status != 0 {
			return fmt.Errorf("link init: dev:%s hotplug add failed (%w)", params.devName, err(status))
		}
	}

	// get port id
	if params.devName != "" {
		port_id, res = lled.GetPortByName(params.devName)
		if res != nil {
			return res
		}
	} else {
		port_id = lled.Port(params.portId)
		if !port_id.IsValid() {
			return errors.New("link init: no valid port id")
		}
	}

	// get device information
	res = port_id.InfoGet(&port_info)
	if res != nil {
		return res
	}

	// check requested receive RSS parameters for this device
	rss = params.rx.rss
	if rss != nil {
		if port_info.RetaSize() == 0 || port_info.RetaSize() > C.RTE_ETH_RSS_RETA_SIZE_512 {
			return errors.New("link init: ethdev redirection table size is 0 or too large (>512)")
		}

		if rss.n_queues == 0 || rss.n_queues >= LINK_RXQ_RSS_MAX {
			return errors.New("link init: requested # queues for RSS is 0 or too large (>16)")
		}

		maxRxQueues := (uint32)(port_info.MaxRxQueues())
		for i := 0; uint32(i) < rss.n_queues; i++ {
			if rss.queue_id[i] >= maxRxQueues {
				return errors.New("link init: requested queue id > ethdev maximum # of Rx queues")
			}
		}
	}

	//
	// Port Resource create
	//

	// configure port config attributes to new port config
	var mtu uint32 = 9000 - (C.RTE_ETHER_HDR_LEN + C.RTE_ETHER_CRC_LEN)
	if params.rx.mtu > 0 {
		mtu = params.rx.mtu
	}

	var optRss = lled.OptRss(lled.RssConf{})
	var rxMqMode = C.RTE_ETH_MQ_RX_NONE
	if rss != nil {
		rxMqMode = C.RTE_ETH_MQ_RX_RSS
		optRss = lled.OptRss(lled.RssConf{
			Hf: (C.RTE_ETH_RSS_IP | C.RTE_ETH_RSS_TCP | C.RTE_ETH_RSS_UDP) & port_info.FlowTypeRssOffloads(),
		})
	}

	res = port_id.DevConfigure(params.rx.n_queues, params.tx.n_queues,
		lled.OptLinkSpeeds(0),
		lled.OptRxMode(lled.RxMode{MqMode: uint(rxMqMode), MTU: mtu, SplitHdrSize: 0}),
		lled.OptTxMode(lled.TxMode{MqMode: C.RTE_ETH_MQ_TX_NONE}),
		optRss,
		lled.OptLoopbackMode(0),
	)
	if res != nil {
		return err(status)
	}

	// if requested set deviceport to promiscuous mode
	if params.promiscuous {
		res = port_id.PromiscEnable()
		if res != nil {
			return res
		}
	}

	cpu_id := port_id.SocketID()
	if cpu_id == C.SOCKET_ID_ANY {
		cpu_id = 0
	}

	// Port RX queues setup
	for i := 0; uint16(i) < params.rx.n_queues; i++ {
		status = C.rte_eth_rx_queue_setup(
			(C.ushort)(port_id), (C.ushort)(i), (C.ushort)(params.rx.queue_size), (C.uint)(cpu_id), nil, mempool.m,
		)
		if status < 0 {
			return err(status)
		}
	}

	// Port TX queues setup
	for i := 0; uint16(i) < params.tx.n_queues; i++ {
		status = C.rte_eth_tx_queue_setup(
			(C.ushort)(port_id), (C.ushort)(i), (C.ushort)(params.tx.queue_size), (C.uint)(cpu_id), nil,
		)
		if status < 0 {
			return err(status)
		}
	}

	// Port start
	res = port_id.Start()
	if res != nil {
		return res
	}

	// configure device rss (receive side scaling) settings
	if rss != nil {
		status = (C.int)(rss_setup(port_id, port_info.RetaSize(), rss))
		if status != 0 {
			port_id.Stop()
			return err(status)
		}
	}

	// Port link up
	res = port_id.SetLinkUp()
	if res != nil { //&& (res != -C.ENOTSUP) {
		port_id.Stop()
		return res
	}

	// Node fill in
	ethdev.name = name
	ethdev.portId = port_id
	ethdev.dev_name, res = port_id.Name()
	if res != nil {
		return res
	}
	ethdev.nRxQ = params.rx.n_queues
	ethdev.nTxQ = params.tx.n_queues
	ethdev.clean = clean

	return nil
}

// Free deletes the current Ethdev record and calls the clean callback function given at init
func (ethdev *Ethdev) Free() {
	// Release all resources for this port
	ethdev.portId.Stop()

	// call given clean callback function if given during init
	if ethdev.clean != nil {
		ethdev.clean()
	}
}

// TODO Rewrite this function to portId.RssRetaUpdate!
func rss_setup(portId lled.Port, reta_size uint16, rss *EthdevParamsRss) int {
	var reta_conf [RETA_CONF_SIZE]C.struct_rte_eth_rss_reta_entry64
	var i uint16
	var status int

	// RETA setting
	for i = 0; i < reta_size; i++ {
		reta_conf[i/C.RTE_ETH_RETA_GROUP_SIZE].mask = C.UINT64_MAX
	}

	for i = 0; i < reta_size; i++ {
		reta_id := (C.uint32_t)(i / C.RTE_ETH_RETA_GROUP_SIZE)
		reta_pos := (C.uint32_t)(i % C.RTE_ETH_RETA_GROUP_SIZE)
		rss_qs_pos := (C.uint32_t)(i % (uint16)(rss.n_queues))

		reta_conf[reta_id].reta[reta_pos] = (C.uint16_t)(rss.queue_id[rss_qs_pos]) //uint16 type?
	}

	//portId.RssRetaUpdate(([]ethdev.RssRetaEntry64)(reta_conf), reta_size)
	// RETA update
	status = (int)(C.rte_eth_dev_rss_reta_update((C.ushort)(portId), (*C.struct_rte_eth_rss_reta_entry64)(&reta_conf[0]),
		(C.ushort)(reta_size)))
	return status
}

func (ethdev *Ethdev) IsUp() (bool, error) {
	linkParams, result := ethdev.portId.EthLinkGet()
	if result != nil {
		return false, result
	}

	return linkParams.Status(), nil
}
