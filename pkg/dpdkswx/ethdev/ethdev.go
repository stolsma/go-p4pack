// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package ethdev

/*
#include <stdlib.h>
#include <stdint.h>
#include <string.h>

#include <net/if.h>

#include <rte_ethdev.h>

*/
import "C"
import (
	"errors"
	"fmt"
	"syscall"
	"unsafe"

	"github.com/stolsma/go-p4pack/pkg/dpdkswx/common"
	lled "github.com/yerden/go-dpdk/ethdev"
	"github.com/yerden/go-dpdk/mempool"
)

const RetaConfSize = (C.RTE_ETH_RSS_RETA_SIZE_512 / C.RTE_ETH_RETA_GROUP_SIZE)

const LinkRxqRssMax = 16

type ParamsRss struct {
	queueID [LinkRxqRssMax]uint32
	nQueues uint32
}

type LinkParams struct {
	DevName           string
	DevArgs           string
	DevHotplugEnabled bool

	Rx struct {
		Mtu       uint32
		NQueues   uint16
		QueueSize uint32
		Mempool   *mempool.Mempool
		Rss       *ParamsRss
	}
	Tx struct {
		NQueues   uint16
		QueueSize uint32
	}

	Promiscuous bool
}

// Ethdev represents a Ethdev record
type Ethdev struct {
	name    string
	devName string
	portID  lled.Port
	nRxQ    uint16
	nTxQ    uint16
	clean   func()
}

// Create Ethdev interface. Returns error when something went wrong.
func (ethdev *Ethdev) Init(name string, params *LinkParams, clean func()) error {
	var portInfo DevInfo
	var rss *ParamsRss // *C.struct_link_params_rss
	var status C.int
	var res error

	// TODO add all params to check!
	// Check input params
	if (name == "") || (params == nil) || (params.Rx.NQueues == 0) ||
		(params.Rx.QueueSize == 0) || (params.Tx.NQueues == 0) || (params.Tx.QueueSize == 0) {
		return nil
	}

	//	cname := C.CString(name)
	//	defer C.free(unsafe.Pointer(cname))
	devName := C.CString(params.DevName)
	defer C.free(unsafe.Pointer(devName))
	devArgs := C.CString(params.DevArgs)
	defer C.free(unsafe.Pointer(devArgs))

	// Performing Device Hotplug and valid for only VDEVs
	if params.DevHotplugEnabled {
		vdev := C.CString("vdev")
		status = C.rte_eal_hotplug_add(vdev, devName, devArgs)
		C.free(unsafe.Pointer(vdev))
		if status != 0 {
			return fmt.Errorf("link init: dev:%s hotplug add failed (%w)", params.DevName, common.Err(status))
		}
	}

	// get port id
	portID, res := lled.GetPortByName(params.DevName)
	if res != nil {
		return res
	}

	// get device information
	// TODO Get added methods upstreamed to go-dpdk!
	res = InfoGet(portID, &portInfo)
	if res != nil {
		return res
	}

	// check requested receive RSS parameters for this device
	rss = params.Rx.Rss
	if rss != nil {
		if portInfo.RetaSize() == 0 || portInfo.RetaSize() > C.RTE_ETH_RSS_RETA_SIZE_512 {
			return errors.New("link init: ethdev redirection table size is 0 or too large (>512)")
		}

		if rss.nQueues == 0 || rss.nQueues >= LinkRxqRssMax {
			return errors.New("link init: requested # queues for RSS is 0 or too large (>16)")
		}

		//		maxRxQueues := (uint32)(portInfo.MaxRxQueues())
		//		for i := 0; uint32(i) < rss.nQueues; i++ {
		//			if rss.queueID[i] >= maxRxQueues {
		//				return errors.New("link init: requested queue id > ethdev maximum # of Rx queues")
		//			}
		//		}
	}

	//
	// Port Resource create
	//

	// configure port config attributes to new port config
	var mtu uint32 = 9000 - (C.RTE_ETHER_HDR_LEN + C.RTE_ETHER_CRC_LEN)
	if params.Rx.Mtu > 0 {
		mtu = params.Rx.Mtu
	}

	var optRss = lled.OptRss(lled.RssConf{})
	var rxMqMode = C.RTE_ETH_MQ_RX_NONE
	if rss != nil {
		rxMqMode = C.RTE_ETH_MQ_RX_RSS
		optRss = lled.OptRss(lled.RssConf{
			Hf: (C.RTE_ETH_RSS_IP | C.RTE_ETH_RSS_TCP | C.RTE_ETH_RSS_UDP) & portInfo.FlowTypeRssOffloads(),
		})
	}

	res = portID.DevConfigure(params.Tx.NQueues, params.Tx.NQueues,
		lled.OptLinkSpeeds(0),
		lled.OptRxMode(lled.RxMode{MqMode: uint(rxMqMode), MTU: mtu, SplitHdrSize: 0}),
		lled.OptTxMode(lled.TxMode{MqMode: C.RTE_ETH_MQ_TX_NONE}),
		optRss,
		lled.OptLoopbackMode(0),
	)
	if res != nil {
		return common.Err(status)
	}

	// if requested set deviceport to promiscuous mode
	if params.Promiscuous {
		res = portID.PromiscEnable()
		if res != nil {
			return res
		}
	}

	cpuID := portID.SocketID()
	if cpuID == C.SOCKET_ID_ANY {
		cpuID = 0
	}

	// Port RX queues setup
	for i := 0; uint16(i) < params.Rx.NQueues; i++ {
		status = C.rte_eth_rx_queue_setup(
			(C.ushort)(portID),
			(C.ushort)(i),
			(C.ushort)(params.Rx.QueueSize),
			(C.uint)(cpuID),
			nil,
			(*C.struct_rte_mempool)(unsafe.Pointer(params.Rx.Mempool)),
		)
		if status < 0 {
			return common.Err(status)
		}
	}

	// Port TX queues setup
	for i := 0; uint16(i) < params.Tx.NQueues; i++ {
		status = C.rte_eth_tx_queue_setup(
			(C.ushort)(portID), (C.ushort)(i), (C.ushort)(params.Tx.QueueSize), (C.uint)(cpuID), nil,
		)
		if status < 0 {
			return common.Err(status)
		}
	}

	// Port start
	res = portID.Start()
	if res != nil {
		return res
	}

	// configure device rss (receive side scaling) settings
	if rss != nil {
		status = (C.int)(rssSetup(portID, portInfo.RetaSize(), rss))
		if status != 0 {
			portID.Stop()
			return common.Err(status)
		}
	}

	// Port link up
	res = portID.SetLinkUp()
	if res != nil && !errors.Is(res, syscall.ENOTSUP) {
		portID.Stop()
		return res
	}

	// Node fill in
	ethdev.name = name
	ethdev.portID = portID
	ethdev.devName, res = portID.Name()
	if res != nil {
		return res
	}
	ethdev.nRxQ = params.Rx.NQueues
	ethdev.nTxQ = params.Tx.NQueues
	ethdev.clean = clean

	return nil
}

func (ethdev *Ethdev) DevName() string {
	return ethdev.devName
}

// Free deletes the current Ethdev record and calls the clean callback function given at init
func (ethdev *Ethdev) Free() {
	// Release all resources for this port
	ethdev.portID.Stop()

	// call given clean callback function if given during init
	if ethdev.clean != nil {
		ethdev.clean()
	}
}

// TODO Rewrite this function to portId.RssRetaUpdate!
func rssSetup(portID lled.Port, retaSize uint16, rss *ParamsRss) int {
	var retaConf [RetaConfSize]C.struct_rte_eth_rss_reta_entry64
	var i uint16
	var status int

	// RETA setting
	for i = 0; i < retaSize; i++ {
		retaConf[i/C.RTE_ETH_RETA_GROUP_SIZE].mask = C.UINT64_MAX
	}

	for i = 0; i < retaSize; i++ {
		retaID := (C.uint32_t)(i / C.RTE_ETH_RETA_GROUP_SIZE)
		retaPos := (C.uint32_t)(i % C.RTE_ETH_RETA_GROUP_SIZE)
		rssQsPos := (C.uint32_t)(i % (uint16)(rss.nQueues))

		retaConf[retaID].reta[retaPos] = (C.uint16_t)(rss.queueID[rssQsPos]) //  uint16 type?
	}

	// portId.RssRetaUpdate(([]ethdev.RssRetaEntry64)(reta_conf), reta_size)
	// RETA update
	status = (int)(C.rte_eth_dev_rss_reta_update((C.ushort)(portID), &retaConf[0], (C.ushort)(retaSize)))
	return status
}

func (ethdev *Ethdev) IsUp() (bool, error) {
	linkParams, result := ethdev.portID.EthLinkGet()
	if result != nil {
		return false, result
	}

	return linkParams.Status(), nil
}

// DevInfo is a structure used to retrieve the contextual information
// of an Ethernet device, such as the controlling driver of the
// device, etc...
type DevInfo C.struct_rte_eth_dev_info

// DriverName returns driver_name as a Go string.
func (info *DevInfo) DriverName() string {
	return C.GoString((*C.struct_rte_eth_dev_info)(info).driver_name)
}

// InterfaceName is the name of the interface in the system.
func (info *DevInfo) InterfaceName() string {
	var buf [C.IF_NAMESIZE]C.char
	return C.GoString(C.if_indextoname(info.if_index, &buf[0]))
}

// RetaSize returns Device redirection table size, the total number of
// entries.
func (info *DevInfo) RetaSize() uint16 {
	return uint16(info.reta_size)
}

// MaxRxQueues returns Device maximum Receive queues.
func (info *DevInfo) MaxRxQueues() uint16 {
	return uint16(info.max_rx_queues)
}

// MaxRxQueues returns Device maximum Transmit queues.
func (info *DevInfo) MaxTxQueues() uint16 {
	return uint16(info.max_tx_queues)
}

// MaxRxQueues returns bit mask of RSS offloads, the bit offset also means flow type.
func (info *DevInfo) FlowTypeRssOffloads() uint64 {
	return uint64(info.flow_type_rss_offloads)
}

func InfoGet(pid lled.Port, info *DevInfo) error {
	return common.Err(C.rte_eth_dev_info_get(C.ushort(pid), (*C.struct_rte_eth_dev_info)(info)))
}
