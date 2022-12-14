// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"errors"
	"fmt"

	"github.com/stolsma/go-p4pack/pkg/dpdkinfra"
	"github.com/stolsma/go-p4pack/pkg/dpdkswx/ethdev"
	"github.com/stolsma/go-p4pack/pkg/dpdkswx/netlink"
	"github.com/stolsma/go-p4pack/pkg/dpdkswx/tap"
)

type InterfaceConfig struct {
	Name   string     `json:"name"`
	Tap    *TapParams `json:"tap"`
	Vdev   *PMDParams `json:"vdev"`
	EthDev *PMDParams `json:"ethdev"`
}

func (i *InterfaceConfig) GetName() string {
	return i.Name
}

// TapConfig represents Tap config parameters
type TapParams struct {
	Rx *struct {
		Mtu     int    `json:"mtu"`
		PktMbuf string `json:"pktmbuf"`
	}
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

// Create interfaces with a given interface configuration list
func (c *Config) ApplyInterface() error {
	dpdki := dpdkinfra.Get()
	if dpdki == nil {
		return errors.New("dpdkinfra module is not initialized")
	}

	// create/bind & configure TAP interface devices
	for _, ifConfig := range c.Interfaces {
		if ifConfig.Tap != nil {
			var tp tap.Params
			name := ifConfig.GetName()

			// get Packet buffer memory pool & MTU
			mpName := ifConfig.Tap.Rx.PktMbuf
			tp.Pktmbuf = dpdki.PktmbufStore.Get(mpName)
			if tp.Pktmbuf == nil {
				return fmt.Errorf("vhost %s mempool %s not found", name, mpName)
			}
			tp.Mtu = ifConfig.Tap.Rx.Mtu

			// create tap
			_, err := dpdki.TapCreate(name, &tp)
			if err != nil {
				return fmt.Errorf("TAP %s create err: %d", name, err)
			}

			// TODO Temporaraly set interface up here but refactor interfaces into seperate dpdki module!
			netlink.InterfaceUp(name)
			netlink.RemoveAllAddr(name)
			log.Infof("TAP %s created!", name)
			continue
		}

		// create and/or bind & configure PMD devices to this environment
		if ifConfig.Vdev != nil || ifConfig.EthDev != nil {
			var vh *PMDParams
			var p ethdev.Params
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
				return fmt.Errorf("vhost %s mempool %s not found", name, vh.Rx.PktMbuf)
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
				return fmt.Errorf("vdev/ethdev %s create err: %d", name, err)
			}

			// TODO Temporaraly set interface up here but refactor interfaces into seperate dpdki module!
			netlink.InterfaceUp(name)
			netlink.RemoveAllAddr(name)
			log.Infof("PMD %s (device name: %s) created!", name, vh.DevName)
		}
	}

	return nil
}
