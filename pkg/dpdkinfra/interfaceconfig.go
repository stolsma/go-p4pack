// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package dpdkinfra

import (
	"github.com/stolsma/go-p4pack/pkg/dpdkswx/ethdev"
	"github.com/vishvananda/netlink"
)

// TapConfig represents Tap config parameters
type TapConfig struct {
}

type InterfaceConfig struct {
	Name   string     `json:"name"`
	Tap    *TapConfig `json:"tap"`
	Vdev   *PMDParams `json:"vdev"`
	EthDev *PMDParams `json:"ethdev"`
}

func (i *InterfaceConfig) GetName() string {
	return i.Name
}

type PMDParams struct {
	DevName string `json:"devname"`
	DevArgs string `json:"devargs"`
	Rx      *struct {
		Mtu         uint32           `json:"mtu"`
		NQueues     uint16           `json:"nqueues"`
		QueueSize   uint32           `json:"queuesize"`
		PktMbuf     string           `json:"pktmbuf"`
		Rss         ethdev.ParamsRss `json:"rss"`
		Promiscuous bool             `json:"promiscuous"`
	}
	Tx *struct {
		NQueues   uint16 `json:"nqueues"`
		QueueSize uint32 `json:"queuesize"`
	}
}

// Create/bind devices (interfaces) through the DpdkInfra API
func (dpdki *DpdkInfra) InterfaceWithConfig(ifConfig *InterfaceConfig) {
	// create/bind & configure TAP interface devices
	if ifConfig.Tap != nil {
		name := ifConfig.GetName()
		_, err := dpdki.TapCreate(name)
		if err != nil {
			log.Fatalf("TAP %s create err: %d", name, err)
		}

		// TODO Temporaraly set interface up here but refactor interfaces into seperate dpdki module!
		dpdki.InterfaceUp(name)
		log.Infof("TAP %s created!", name)
		return
	}

	// create and/or bind & configure PMD devices to this environment
	if ifConfig.Vdev != nil || ifConfig.EthDev != nil {
		var vh *PMDParams
		var p ethdev.LinkParams

		name := ifConfig.GetName()
		if ifConfig.Vdev != nil {
			vh = ifConfig.Vdev
			p.DevHotplugEnabled = true
		} else {
			vh = ifConfig.EthDev
		}

		// get Packet buffer memory pool
		mp := dpdki.PktmbufStore.Get(vh.Rx.PktMbuf)
		if mp == nil {
			log.Fatalf("vhost %s mempool %s not found", name, vh.Rx.PktMbuf)
		}

		// copy parameters
		p.DevName = vh.DevName
		p.DevArgs = vh.DevArgs
		p.Rx.Mtu = vh.Rx.Mtu
		p.Rx.NQueues = vh.Rx.NQueues
		p.Rx.QueueSize = vh.Rx.QueueSize
		p.Rx.Mempool = mp.Mempool()
		p.Rx.Rss = vh.Rx.Rss
		p.Tx.NQueues = vh.Tx.NQueues
		p.Tx.QueueSize = vh.Tx.QueueSize
		p.Promiscuous = vh.Rx.Promiscuous

		// create and configure the PMD interface
		_, err := dpdki.EthdevCreate(name, &p)
		if err != nil {
			log.Fatalf("vdev/ethdev %s create err: %d", name, err)
		}

		// TODO Temporaraly set interface up here but refactor interfaces into seperate dpdki module!
		dpdki.InterfaceUp(name)
		log.Infof("PMD %s (device name: %s) created!", name, vh.DevName)
	}
}

// Create interfaces through the DpdkInfra API
func (dpdki *DpdkInfra) InterfaceUp(name string) error {
	// set interface up if not already up
	localInterface, err := netlink.LinkByName(name)
	if err != nil {
		return err
	}
	if err = netlink.LinkSetUp(localInterface); err != nil {
		return err
	}

	list, _ := netlink.AddrList(localInterface, netlink.FAMILY_ALL)
	for _, addr := range list {
		a := addr
		err := netlink.AddrDel(localInterface, &a)
		if err != nil {
			log.Warnf("Couldnt remove address %s from interface %s", addr.String(), name)
		}
	}

	return nil
}
