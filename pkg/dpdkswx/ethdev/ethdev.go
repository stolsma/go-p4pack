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
	DevName string
	Rx      struct {
		Mtu       uint16
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
	port    lled.Port
	devName string
	nRxQ    uint16
	nTxQ    uint16
}

// Initialize ethdev struct
func (ethdev *Ethdev) Init(name string) {
	ethdev.Device = &device.Device{}
	ethdev.SetType("PMD")
	ethdev.SetName(name)
}

// Create and/or configure DPDK Ethdev device. Returns error when something went wrong.
func (ethdev *Ethdev) Initialize(params *Params, clean func()) error {
	var portInfo DevInfo
	var status C.int
	var res error

	// TODO add all params to check!
	// Check input params
	if (params == nil) || (params.Rx.NQueues == 0) ||
		(params.Rx.QueueSize == 0) || (params.Tx.NQueues == 0) || (params.Tx.QueueSize == 0) {
		return nil
	}

	// get port id and save to this struct!
	portID, res := lled.GetPortByName(params.DevName)
	if res != nil {
		return res
	}
	ethdev.port = portID

	// get ethDev device information
	res = ethdev.InfoGet(&portInfo)
	if res != nil {
		return res
	}

	// check requested MTU value
	var mtu = portInfo.MaxMTU()
	if params.Rx.Mtu > 0 {
		if params.Rx.Mtu >= portInfo.MinMTU() && params.Rx.Mtu <= portInfo.MaxMTU() {
			mtu = params.Rx.Mtu
		} else {
			return errors.New("requested MTU is smaller than minimum MTU or larger then maximum MTU supported for this port")
		}
	}

	// check maximum number of queues to configure to the max supported queues on device
	if params.Rx.NQueues > portInfo.MaxRxQueues() || params.Tx.NQueues > portInfo.MaxTxQueues() {
		return errors.New("number of Tx or Rx queues to large")
	}

	// check requested receive RSS parameters for this device
	rss := params.Rx.Rss
	if rss != nil {
		if portInfo.RetaSize() == 0 || portInfo.RetaSize() > EthRssRetaSize512 {
			return errors.New("ethdev redirection table size (rss) is 0 or too large (>512)")
		}

		for i := 0; i < len(rss); i++ {
			if rss[i] >= params.Rx.NQueues {
				return errors.New("the RSS queue id > maximum requested # of Rx queues")
			}
		}
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
	res = portID.DevConfigure(params.Rx.NQueues, params.Tx.NQueues,
		lled.OptLinkSpeeds(0),
		lled.OptRxMode(lled.RxMode{MqMode: uint(rxMqMode), MTU: uint32(mtu), SplitHdrSize: 0}),
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
			log.Infof("PMD %s does not support promiscuous mode", ethdev.Name())
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
		log.Infof("PMD %s does not support SetLinkUp", ethdev.Name())
	}

	// Node fill in
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

func (ethdev *Ethdev) SamePort(ethdev2 *Ethdev) bool {
	return ethdev.port == ethdev2.port
}

// Free deletes the current Ethdev record and calls the clean callback function given at init
func (ethdev *Ethdev) Free() error {
	// Release all resources for this port
	ethdev.port.Stop()

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
	linkParams, result := ethdev.port.EthLinkGet()
	if result != nil {
		return false, result
	}

	return linkParams.Status(), nil
}

func (ethdev *Ethdev) SetLinkUp() error {
	err := ethdev.port.SetLinkUp()
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
	err := ethdev.port.SetLinkDown()
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
	err := ethdev.port.StatsGet(&stats)
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

	linkParams, err := ethdev.port.EthLinkGet()
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
	err = ethdev.port.MACAddrGet(addr)
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

// Get list of attached ethdev ports (look out: an Ethdev device could have multiple ports!)
func GetAttachedPorts() ([]*Ethdev, error) {
	var ports []*Ethdev
	attachedPorts := lled.ValidPorts()
	for _, port := range attachedPorts {
		var e = Ethdev{}

		name, err := port.Name()
		if err != nil {
			return nil, fmt.Errorf("error reading port name: %v", err)
		}

		// set ethdev struct
		e.Init(name)
		e.devName = name
		e.port = port
		ports = append(ports, &e)
	}

	return ports, nil
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

type RteEthDevFlags uint32

// Device flags, flags internally saved in rte_eth_dev_data.dev_flags and reported in rte_eth_dev_info.dev_flags.
const (
	// PMD supports thread-safe flow operations
	RteEthDevFlowOpsThreadSafe RteEthDevFlags = C.RTE_ETH_DEV_FLOW_OPS_THREAD_SAFE
	// Device supports link state interrupt
	RteEthDevIntrLsc = C.RTE_ETH_DEV_INTR_LSC
	// Device is a bonded slave
	RteEthDevBondedSlave = C.RTE_ETH_DEV_BONDED_SLAVE
	// Device supports device removal interrupt
	RteEthDevIntrRmv = C.RTE_ETH_DEV_INTR_RMV
	// Device is port representor
	RteEthDevRepresentor = C.RTE_ETH_DEV_REPRESENTOR
	// Device does not support MAC change after started
	RteEthDevNoliveMACAddr = C.RTE_ETH_DEV_NOLIVE_MAC_ADDR
	// Queue xstats filled automatically by ethdev layer. PMDs filling the queue xstats themselves should not set this flag
	RteEthDevAutofillQueueXstats = C.RTE_ETH_DEV_AUTOFILL_QUEUE_XSTATS
)

// Device flags.
func (info *DevInfo) DeviceFlags() RteEthDevFlags {
	return RteEthDevFlags(*info.dev_flags)
}

// RetaSize returns Device redirection table size, the total number of entries.
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
	return uint16(info.max_mtu)
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

/*
	TODO Following fields need to be added as retrieval function:
	min_rx_bufsize         _Ctype_uint32_t
	max_rx_pktlen          _Ctype_uint32_t
	max_lro_pkt_size       _Ctype_uint32_t
	max_mac_addrs          _Ctype_uint32_t
	max_hash_mac_addrs     _Ctype_uint32_t
	max_vfs                _Ctype_uint16_t
	max_vmdq_pools         _Ctype_uint16_t
	rx_seg_capa            _Ctype_struct_rte_eth_rxseg_capa
	rx_offload_capa        _Ctype_uint64_t
	tx_offload_capa        _Ctype_uint64_t
	rx_queue_offload_capa  _Ctype_uint64_t
	tx_queue_offload_capa  _Ctype_uint64_t
	hash_key_size          _Ctype_uint8_t
	flow_type_rss_offloads _Ctype_uint64_t
	default_rxconf         _Ctype_struct_rte_eth_rxconf
	default_txconf         _Ctype_struct_rte_eth_txconf
	vmdq_queue_base        _Ctype_uint16_t
	vmdq_queue_num         _Ctype_uint16_t
	vmdq_pool_base         _Ctype_uint16_t
	rx_desc_lim            _Ctype_struct_rte_eth_desc_lim
	tx_desc_lim            _Ctype_struct_rte_eth_desc_lim
	speed_capa             _Ctype_uint32_t
	nb_rx_queues           _Ctype_uint16_t
	nb_tx_queues           _Ctype_uint16_t
	default_rxportconf     _Ctype_struct_rte_eth_dev_portconf
	default_txportconf     _Ctype_struct_rte_eth_dev_portconf
	dev_capa               _Ctype_uint64_t
	switch_info            _Ctype_struct_rte_eth_switch_info
*/

//
// Extra Ethdev methods to be upstreamed
//

func (ethdev *Ethdev) InfoGet(info *DevInfo) error {
	return common.Err(C.rte_eth_dev_info_get(C.ushort(ethdev.port), (*C.struct_rte_eth_dev_info)(info)))
}

var PromiscuousModeStr = [2]string{0: "off", 1: "on"}

func (ethdev *Ethdev) PromiscuousGet() int {
	return int(C.rte_eth_promiscuous_get(C.ushort(ethdev.port)))
}

func RteEthDevTxOffloadName(txOffload uint64) string {
	// no free needed, returned C string is static!
	cTxOffloadName := C.rte_eth_dev_tx_offload_name(C.uint64_t(txOffload))
	return C.GoString(cTxOffloadName)
}

func RteEthDevRxOffloadName(rxOffload uint64) string {
	// no free needed, returned C string is static!
	cRxOffloadName := C.rte_eth_dev_rx_offload_name(C.uint64_t(rxOffload))
	return C.GoString(cRxOffloadName)
}

func RteEthDevCapabilityName(capability uint64) string {
	// no free needed, returned C string is static!
	cCapName := C.rte_eth_dev_capability_name(C.uint64_t(capability))
	return C.GoString(cCapName)
}

func RteEthLinkSpeedToString(linkSpeed uint32) string {
	// no free needed, returned C string is static!
	cLinkSpeedName := C.rte_eth_link_speed_to_str(C.uint32_t(linkSpeed))
	return C.GoString(cLinkSpeedName)
}
