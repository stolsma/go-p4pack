// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package ethdev

/*
#include <stdlib.h>
#include <stdint.h>
#include <string.h>

#include <net/if.h>

#include <rte_ethdev.h>
#include <rte_swx_port_ethdev.h>

*/
import "C"
import (
	"errors"
	"fmt"
	"syscall"
	"unsafe"

	"github.com/stolsma/go-p4pack/pkg/dpdkswx/common"
	"github.com/stolsma/go-p4pack/pkg/dpdkswx/device"
	"github.com/stolsma/go-p4pack/pkg/dpdkswx/pipeline"
	"github.com/stolsma/go-p4pack/pkg/dpdkswx/pktmbuf"
	"github.com/stolsma/go-p4pack/pkg/logging"
	lled "github.com/yerden/go-dpdk/ethdev"
)

var log logging.Logger

func init() {
	// keep the logger up to date, also after new log config
	logging.Register("dpdkswx/ethdev", func(logger logging.Logger) {
		log = logger
	})
}

const RetaConfSize = (EthRssRetaSize512 / EthRetaGroupSize)

type ParamsRss []uint16

type Params struct {
	DevName           string
	DevArgs           string
	DevHotplugEnabled bool

	Rx struct {
		Mtu       uint32
		NQueues   uint16
		QueueSize uint32
		Mempool   *pktmbuf.Pktmbuf
		Rss       ParamsRss
	}
	Tx struct {
		NQueues   uint16
		QueueSize uint32
	}

	Promiscuous bool
}

// Ethdev represents a Ethdev record
type Ethdev struct {
	*device.Device
	*lled.Port
	devName string
	nRxQ    uint16
	nTxQ    uint16
}

// Create and/or configure DPDK Ethdev device. Returns error when something went wrong.
func (ethdev *Ethdev) Init(name string, params *Params, clean func()) error {
	var portInfo DevInfo
	var status C.int
	var res error

	// TODO add all params to check!
	// Check input params
	if (name == "") || (params == nil) || (params.Rx.NQueues == 0) ||
		(params.Rx.QueueSize == 0) || (params.Tx.NQueues == 0) || (params.Tx.QueueSize == 0) {
		return nil
	}

	cDevName := C.CString(params.DevName)
	defer C.free(unsafe.Pointer(cDevName))
	cDevArgs := C.CString(params.DevArgs)
	defer C.free(unsafe.Pointer(cDevArgs))

	// Performing Device Hotplug and valid for only VDEVs
	if params.DevHotplugEnabled {
		vdev := C.CString("vdev")
		status = C.rte_eal_hotplug_add(vdev, cDevName, cDevArgs)
		C.free(unsafe.Pointer(vdev))
		if status != 0 {
			return fmt.Errorf("link init: dev:%s hotplug add failed (%w)", params.DevName, common.Err(status))
		}
	}

	// get port id and save to this struct!
	portID, res := lled.GetPortByName(params.DevName)
	if res != nil {
		return res
	}
	ethdev.Port = &portID

	// get ethDev device information
	res = ethdev.InfoGet(&portInfo)
	if res != nil {
		return res
	}

	// check maximum number of queues to configure to the max supported queues on device
	if params.Rx.NQueues > portInfo.MaxRxQueues() || params.Rx.NQueues > portInfo.MaxTxQueues() {
		return errors.New("link init: Number of Tx or Rx queues to large")
	}

	// check requested receive RSS parameters for this device
	rss := params.Rx.Rss
	if rss != nil {
		if portInfo.RetaSize() == 0 || portInfo.RetaSize() > EthRssRetaSize512 {
			return errors.New("link init: ethdev redirection table size is 0 or too large (>512)")
		}

		for i := 0; i < len(rss); i++ {
			if rss[i] >= params.Rx.NQueues {
				return errors.New("link init: RSS queue id > maximum requested # of Rx queues")
			}
		}
	}

	// configure port config attributes to new port config
	// TODO Check if device supports it!
	var mtu uint32 = 9000 - (EtherHdrLen + EtherCRCLen)
	if params.Rx.Mtu > 0 {
		mtu = params.Rx.Mtu
	}

	// define device rss parameters
	var optRss = lled.OptRss(lled.RssConf{})
	var rxMqMode = EthMqRxNone
	if rss != nil {
		rxMqMode = EthMqRxRss
		optRss = lled.OptRss(lled.RssConf{
			Hf: (EthRssIP | EthRssTCP | EthRssUDP) & portInfo.FlowTypeRssOffloads(),
		})
	}

	// configure the ethdev device
	res = portID.DevConfigure(params.Tx.NQueues, params.Tx.NQueues,
		lled.OptLinkSpeeds(0),
		lled.OptRxMode(lled.RxMode{MqMode: uint(rxMqMode), MTU: mtu, SplitHdrSize: 0}),
		lled.OptTxMode(lled.TxMode{MqMode: EthMqTxNone}),
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
			if !errors.Is(res, syscall.ENOTSUP) {
				return res
			}
			log.Infof("PMD %s does not support promiscuous mode", name)
		}
	}

	// is the ethdev device connected to a specific CPU Socket?
	cpuID := portID.SocketID()
	if cpuID == C.SOCKET_ID_ANY {
		cpuID = 0
	}

	// Device RX queues setup
	for i := 0; uint16(i) < params.Rx.NQueues; i++ {
		status = C.rte_eth_rx_queue_setup(
			(C.ushort)(portID),
			(C.ushort)(i),
			(C.ushort)(params.Rx.QueueSize),
			(C.uint)(cpuID),
			nil,
			(*C.struct_rte_mempool)(unsafe.Pointer(params.Rx.Mempool.Mempool())),
		)
		if status < 0 {
			return common.Err(status)
		}
	}

	// Device TX queues setup
	for i := 0; uint16(i) < params.Tx.NQueues; i++ {
		status = C.rte_eth_tx_queue_setup(
			(C.ushort)(portID), (C.ushort)(i), (C.ushort)(params.Tx.QueueSize), (C.uint)(cpuID), nil,
		)
		if status < 0 {
			return common.Err(status)
		}
	}

	// Device start
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

	// Device link up
	res = portID.SetLinkUp()
	if res != nil {
		if !errors.Is(res, syscall.ENOTSUP) {
			portID.Stop()
			return res
		}
		log.Infof("PMD %s does not support SetLinkUp", name)
	}

	// Node fill in
	ethdev.Device = &device.Device{}
	ethdev.SetType("PMD")
	ethdev.SetName(name)
	ethdev.devName, res = portID.Name()
	if res != nil {
		return res
	}
	ethdev.nRxQ = params.Rx.NQueues
	ethdev.nTxQ = params.Tx.NQueues

	ethdev.SetPipelineOutPort(device.NotBound)
	ethdev.SetPipelineInPort(device.NotBound)
	ethdev.SetClean(clean)

	return nil
}

func (ethdev *Ethdev) Name() string {
	return ethdev.Device.Name()
}

func (ethdev *Ethdev) DevName() string {
	return ethdev.devName
}

// Free deletes the current Ethdev record and calls the clean callback function given at init
func (ethdev *Ethdev) Free() error {
	// Release all resources for this port
	ethdev.Stop()

	// call given clean callback function if given during init
	if ethdev.Clean() != nil {
		ethdev.Clean()()
	}

	return nil
}

// TODO Rewrite this function to portId.RssRetaUpdate!
func rssSetup(portID lled.Port, retaSize uint16, rss ParamsRss) int {
	var retaConf [RetaConfSize]C.struct_rte_eth_rss_reta_entry64
	var i uint16
	var status int

	// RETA setting, ethdev retasize is always a multiple of RTE_ETH_RETA_GROUP_SIZE!
	for i = 0; i < retaSize; i++ {
		retaConf[i/EthRetaGroupSize].mask = C.UINT64_MAX
	}

	for i = 0; i < retaSize; i++ {
		retaID := (C.uint32_t)(i / EthRetaGroupSize)
		retaPos := (C.uint32_t)(i % EthRetaGroupSize)
		rssQsPos := (C.uint32_t)(i % (uint16)(len(rss)))

		retaConf[retaID].reta[retaPos] = (C.uint16_t)(rss[rssQsPos])
	}

	// portId.RssRetaUpdate(([]ethdev.RssRetaEntry64)(reta_conf), reta_size)
	// RETA update
	status = (int)(C.rte_eth_dev_rss_reta_update((C.ushort)(portID), &retaConf[0], (C.ushort)(retaSize)))
	return status
}

// bind to given pipeline input port
func (ethdev *Ethdev) BindToPipelineInputPort(pl *pipeline.Pipeline, portID int, rxq uint, bsz uint) error {
	var params C.struct_rte_swx_port_ethdev_reader_params

	if ethdev.PipelineInPort() != device.NotBound {
		return errors.New("port already bound")
	}
	ethdev.SetPipelineIn(pl.GetName())
	ethdev.SetPipelineInPort(portID)

	params.dev_name = C.CString(ethdev.DevName())
	defer C.free(unsafe.Pointer(params.dev_name))
	params.queue_id = C.ushort(rxq)
	params.burst_size = (C.uint)(bsz)

	return pl.PortInConfig(portID, "ethdev", unsafe.Pointer(&params))
}

// bind to given pipeline output port
func (ethdev *Ethdev) BindToPipelineOutputPort(pl *pipeline.Pipeline, portID int, txq uint, bsz uint) error {
	var params C.struct_rte_swx_port_ethdev_writer_params

	if ethdev.PipelineOutPort() != device.NotBound {
		return errors.New("port already bound")
	}
	ethdev.SetPipelineOut(pl.GetName())
	ethdev.SetPipelineOutPort(portID)

	params.dev_name = C.CString(ethdev.DevName())
	defer C.free(unsafe.Pointer(params.dev_name))
	params.queue_id = C.ushort(txq)
	params.burst_size = (C.uint)(bsz)

	return pl.PortOutConfig(portID, "ethdev", unsafe.Pointer(&params))
}

func (ethdev *Ethdev) IsUp() (bool, error) {
	linkParams, result := ethdev.EthLinkGet()
	if result != nil {
		return false, result
	}

	return linkParams.Status(), nil
}

func (ethdev *Ethdev) SetLinkUp() error {
	err := ethdev.Port.SetLinkUp()
	if err != nil {
		if !errors.Is(err, syscall.ENOTSUP) {
			return err
		}
		log.Debugf("PMD %v does not support LinkUp operation trying kernel", ethdev.Name())
		// TODO implement try netlink portup!!
	}

	return nil
}

func (ethdev *Ethdev) SetLinkDown() error {
	err := ethdev.Port.SetLinkDown()
	if err != nil {
		if !errors.Is(err, syscall.ENOTSUP) {
			return err
		}
		log.Debugf("PMD %v does not support LinkDown operation trying kernel", ethdev.Name())
		// TODO implement try netlink portdown!!
	}

	return nil
}

func (ethdev *Ethdev) GetPortStatsString() (string, error) {
	var stats lled.Stats
	var info string
	err := ethdev.Port.StatsGet(&stats)
	if err != nil {
		return info, err
	}

	goStats := stats.Cast()

	info += fmt.Sprintf("\tRX packets: %-20d bytes : %-20d\n", goStats.Ipackets, goStats.Ibytes)
	info += fmt.Sprintf("\tRX errors : %-20d missed: %-20d RX no mbuf: %-20d\n", goStats.Ierrors, goStats.Imissed, goStats.RxNoMbuf)
	info += fmt.Sprintf("\tTX packets: %-20d bytes : %-20d\n", goStats.Opackets, goStats.Obytes)
	info += fmt.Sprintf("\tTX errors : %-20d\n", goStats.Oerrors)

	return info, nil
}

func (ethdev *Ethdev) GetPortInfoString() (string, error) {
	var portInfo DevInfo
	info := ""

	linkParams, err := ethdev.EthLinkGet()
	if err != nil {
		return "", err
	}
	if linkParams.Status() {
		info += fmt.Sprintf("  Link Status            : %s \n", "Up")
	} else {
		info += fmt.Sprintf("  Link Status            : %s \n", "Down")
	}
	if linkParams.AutoNeg() {
		info += fmt.Sprintf("  Autonegotioation       : %s \n", "On")
	} else {
		info += fmt.Sprintf("  Autonegotioation       : %s \n", "Off")
	}
	if linkParams.Duplex() {
		info += fmt.Sprintf("  Duplex                 : %s \n", "Full")
	} else {
		info += fmt.Sprintf("  Duplex                 : %s \n", "Half")
	}
	info += fmt.Sprintf("  Linkspeed (mbps)       : %d \n", linkParams.Speed())
	info += "\n"

	info += fmt.Sprintf("  Promiscuous mode       : %s \n", PromiscuousModeStr[ethdev.PromiscuousGet()])
	info += "\n"

	var addr = &lled.MACAddr{}
	err = ethdev.MACAddrGet(addr)
	if err == nil {
		info += fmt.Sprintf("  MAC Address            : %s \n", addr.String())
	}

	err = ethdev.InfoGet(&portInfo)
	if err != nil {
		return "", err
	}
	info += portInfo.String()

	return info, nil
}

/****************************************************************************************
 * Everything defined below this line is missing in go-dpdk and needs to be upstreamed! *
 ****************************************************************************************/

const (
	EtherHdrLen = C.RTE_ETHER_HDR_LEN
	EtherCRCLen = C.RTE_ETHER_CRC_LEN

	EthMqRxNone = C.RTE_ETH_MQ_RX_NONE
	EthMqRxRss  = C.RTE_ETH_MQ_RX_RSS
	EthMqTxNone = C.RTE_ETH_MQ_TX_NONE

	EthRssIP  = C.RTE_ETH_RSS_IP
	EthRssTCP = C.RTE_ETH_RSS_TCP
	EthRssUDP = C.RTE_ETH_RSS_UDP

	EthRetaGroupSize  = C.RTE_ETH_RETA_GROUP_SIZE
	EthRssRetaSize512 = C.RTE_ETH_RSS_RETA_SIZE_512
)

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

// MinMTU returns Device minimum supported MTU size.
func (info *DevInfo) MinMTU() uint16 {
	return uint16(info.min_mtu)
}

// MaxMTU returns Device maximum supported MTU size.
func (info *DevInfo) MaxMTU() uint16 {
	return uint16(info.min_mtu)
}

func (info *DevInfo) String() string {
	result := ""
	result += fmt.Sprintf("  device name            : %s \n", C.GoString(info.device.name))
	result += fmt.Sprintf("  driver name            : %s \n", C.GoString(info.device.driver.name))
	result += fmt.Sprintf("  device alias           : %s \n", C.GoString(info.device.driver.alias))
	result += fmt.Sprintf("  bus                    : %s \n", C.GoString(info.device.bus.name))
	result += fmt.Sprintf("  numa node              : %d \n", info.device.numa_node)
	result += fmt.Sprintf("  ifIndex                : %d \n", info.if_index)
	result += fmt.Sprintf("  min MTU                : %d \n", info.min_mtu)
	result += fmt.Sprintf("  max MTU                : %d \n", info.max_mtu)
	result += fmt.Sprintf("  dev_flags              : %d \n", info.dev_flags)
	result += fmt.Sprintf("  min_rx_bufsize         : %d \n", info.min_rx_bufsize)
	result += fmt.Sprintf("  max_rx_pktlen          : %d \n", info.max_rx_pktlen)
	result += fmt.Sprintf("  max_lro_pkt_size       : %d \n", info.max_lro_pkt_size)
	result += fmt.Sprintf("  max_rx_queue           : %d \n", info.max_rx_queues)
	result += fmt.Sprintf("  max_tx_queue           : %d \n", info.max_tx_queues)
	result += fmt.Sprintf("  max_mac_addrs          : %d \n", info.max_mac_addrs)
	result += fmt.Sprintf("  max_hash_mac_addrs     : %d \n", info.max_hash_mac_addrs)
	result += fmt.Sprintf("  max_vf                 : %d \n", info.max_vfs)
	result += fmt.Sprintf("  max_vmdq_pools         : %d \n", info.max_vmdq_pools)
	// rx_seg_capa		_Ctype_struct_rte_eth_rxseg_capa
	result += fmt.Sprintf("  rx_offload_capa        : %d \n", info.rx_offload_capa)
	result += fmt.Sprintf("  tx_offload_capa        : %d \n", info.tx_offload_capa)
	result += fmt.Sprintf("  rx_queue_offload_capa  : %d \n", info.rx_queue_offload_capa)
	result += fmt.Sprintf("  tx_queue_offload_capa  : %d \n", info.tx_queue_offload_capa)
	result += fmt.Sprintf("  reta_size              : %d \n", info.reta_size)
	result += fmt.Sprintf("  ihash_key_size         : %d \n", info.hash_key_size)
	result += fmt.Sprintf("  flow_type_rss_offloads : %d \n", info.flow_type_rss_offloads)
	result += fmt.Sprintf("  vmdq_queue_base        : %d \n", info.vmdq_queue_base)
	result += fmt.Sprintf("  vmdq_queue_num         : %d \n", info.vmdq_queue_num)
	result += fmt.Sprintf("  vmdq_pool_base         : %d \n", info.vmdq_pool_base)
	// rx_desc_lim		_Ctype_struct_rte_eth_desc_lim
	// tx_desc_lim		_Ctype_struct_rte_eth_desc_lim
	result += fmt.Sprintf("  speed_capa             : %d \n", info.speed_capa)
	result += fmt.Sprintf("  nb_rx_queues           : %d \n", info.nb_rx_queues)
	result += fmt.Sprintf("  nb_tx_queues           : %d \n", info.nb_tx_queues)
	result += fmt.Sprintf("  dev_cap                : %d \n", info.dev_capa)

	return result
}

//
// Extra Ethdev methods to be upstreamed
//

func (ethdev *Ethdev) InfoGet(info *DevInfo) error {
	return common.Err(C.rte_eth_dev_info_get(C.ushort(*ethdev.Port), (*C.struct_rte_eth_dev_info)(info)))
}

var PromiscuousModeStr = [2]string{0: "off", 1: "on"}

func (ethdev *Ethdev) PromiscuousGet() int {
	return int(C.rte_eth_promiscuous_get(C.ushort(*ethdev.Port)))
}
