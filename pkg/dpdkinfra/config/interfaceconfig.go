// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"errors"
	"fmt"

	"github.com/stolsma/go-p4pack/pkg/dpdkinfra"
	"github.com/stolsma/go-p4pack/pkg/dpdkswx/ethdev"
	"github.com/stolsma/go-p4pack/pkg/dpdkswx/netlink"
	"github.com/stolsma/go-p4pack/pkg/dpdkswx/ring"
	"github.com/stolsma/go-p4pack/pkg/dpdkswx/sourcesink"
	"github.com/stolsma/go-p4pack/pkg/dpdkswx/tap"
)

type InterfacesConfig []*InterfaceConfig

type InterfaceConfig struct {
	Name   string        `json:"name"`
	Tap    *TapParams    `json:"tap"`
	EthDev *PMDParams    `json:"ethdev"`
	Ring   *RingParams   `json:"ring"`
	Source *SourceParams `json:"source"`
	Sink   *SinkParams   `json:"sink"`
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

type RingParams struct {
	Size     uint   `json:"size"`
	NumaNode uint32 `json:"numanode"`
}

type SourceParams struct {
	Rx *struct {
		FileName string `json:"filename"`
		NLoops   uint64 `json:"n_loops"`
		NPktsMax uint32 `json:"n_pkts_max"`
		PktMbuf  string `json:"pktmbuf"`
	}
}

type SinkParams struct {
	Tx *struct {
		FileName string `json:"filename"`
	}
}

type PMDParams struct {
	PortName string `json:"portname"`
	Rx       *struct {
		Mtu         uint16           `json:"mtu"`
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
func (c InterfacesConfig) Apply() error {
	dpdki := dpdkinfra.Get()
	if dpdki == nil {
		return errors.New("dpdkinfra module is not initialized")
	}

	// create/bind & configure TAP interface devices
	for _, ifConfig := range c {
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
				return fmt.Errorf("tap %s create err: %d", name, err)
			}

			// TODO Temporaraly set interface up here but refactor interfaces into seperate dpdki module!
			netlink.InterfaceUp(name)
			netlink.RemoveAllAddr(name)
			log.Infof("tap %s created!", name)
			continue
		}

		if ifConfig.Ring != nil {
			var r ring.Params
			name := ifConfig.GetName()

			r.Size = ifConfig.Ring.Size
			r.NumaNode = ifConfig.Ring.NumaNode

			// create ring
			_, err := dpdki.RingCreate(name, &r)
			if err != nil {
				return fmt.Errorf("ring %s create err: %d", name, err)
			}

			log.Infof("ring %s created!", name)
			continue
		}

		if ifConfig.Source != nil {
			var s sourcesink.SourceParams
			name := ifConfig.GetName()

			// get Packet buffer memory pool & all other parameters
			mpName := ifConfig.Source.Rx.PktMbuf
			s.Pktmbuf = dpdki.PktmbufStore.Get(mpName)
			if s.Pktmbuf == nil {
				return fmt.Errorf("source %s mempool %s not found", name, mpName)
			}
			s.FileName = ifConfig.Source.Rx.FileName
			s.NLoops = ifConfig.Source.Rx.NLoops
			s.NPktsMax = ifConfig.Source.Rx.NPktsMax

			// create ring
			_, err := dpdki.SourceCreate(name, &s)
			if err != nil {
				return fmt.Errorf("ring %s create err: %d", name, err)
			}

			log.Infof("Ring %s created!", name)
			continue
		}

		if ifConfig.Sink != nil {
			var s sourcesink.SinkParams
			name := ifConfig.GetName()

			// get parameters
			s.FileName = ifConfig.Sink.Tx.FileName

			// create sink
			_, err := dpdki.SinkCreate(name, &s)
			if err != nil {
				return fmt.Errorf("sink %s create err: %d", name, err)
			}

			log.Infof("sink %s created!", name)
			continue
		}

		// create and/or bind & configure PMD devices to this environment
		if ifConfig.EthDev != nil {
			var vh = ifConfig.EthDev
			var p ethdev.Params
			name := ifConfig.GetName()

			// get Packet buffer memory pool
			mp := dpdki.PktmbufStore.Get(vh.Rx.PktMbuf)
			if mp == nil {
				return fmt.Errorf("vhost %s mempool %s not found", name, vh.Rx.PktMbuf)
			}

			// copy parameters
			p.PortName = vh.PortName
			p.Rx.Mtu = vh.Rx.Mtu
			p.Rx.NQueues = vh.Rx.NQueues
			p.Rx.QueueSize = vh.Rx.QueueSize
			p.Rx.Mempool = mp
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

			log.Infof("Ethdev %s (device port name: %s) created!", name, vh.PortName)
			continue
		}

		log.Errorf("Unknown interface type or wrong configuration for interface %s", ifConfig.GetName())
		return errors.New("error in interface configuration")
	}

	return nil
}
