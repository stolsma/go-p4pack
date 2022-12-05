// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package portmngr

import (
	"errors"
	"fmt"

	"github.com/stolsma/go-p4pack/pkg/dpdkinfra/store"
	"github.com/stolsma/go-p4pack/pkg/dpdkswx/ethdev"
	"github.com/stolsma/go-p4pack/pkg/dpdkswx/ring"
	"github.com/stolsma/go-p4pack/pkg/logging"
)

var log logging.Logger

func init() {
	// keep the logger up to date, also after new log config
	logging.Register("dpdkinfra/portmngr", func(logger logging.Logger) {
		log = logger
	})
}

type PortMngr struct {
	EthdevStore *store.Store[*ethdev.Ethdev]
	RingStore   *store.Store[*ring.Ring]
	TapStore    *store.Store[*Tap]
}

// Initialize the non system intrusive portmngr singleton parts
func (pm *PortMngr) Init() error {
	// create stores
	pm.EthdevStore = store.NewStore[*ethdev.Ethdev]()
	pm.RingStore = store.NewStore[*ring.Ring]()
	pm.TapStore = store.NewStore[*Tap]()

	return nil
}

func (pm *PortMngr) Cleanup() {
	// empty & remove stores
	pm.TapStore.Clear()
	pm.RingStore.Clear()
	pm.EthdevStore.Clear()
}

func (pm *PortMngr) TapCreate(name string) (*Tap, error) {
	var t Tap
	if pm.TapStore.Contains(name) {
		return nil, errors.New("tap with this name exists")
	}

	if err := t.Init(name, func() {
		pm.TapStore.Delete(name)
	}); err != nil {
		return nil, err
	}

	// add node to list
	pm.TapStore.Set(name, &t)
	return &t, nil
}

func (pm *PortMngr) TapList(name string) (string, error) {
	result := ""
	err := pm.TapStore.Iterate(func(key string, tap *Tap) error {
		result += fmt.Sprintf("  %s \n", tap.Name())
		return nil
	})
	return result, err
}

// RingCreate creates a ring and stores it in the portmngr ring store
func (pm *PortMngr) RingCreate(name string, size uint, numaNode uint32) (*ring.Ring, error) {
	var r ring.Ring
	if pm.RingStore.Contains(name) {
		return nil, errors.New("ring with this name exists")
	}

	// initialize
	if err := r.Init(name, size, numaNode, func() {
		pm.RingStore.Delete(name)
	}); err != nil {
		return nil, err
	}

	// add node to list
	pm.RingStore.Set(name, &r)
	log.Infof("ring %s created", name)
	return &r, nil
}

// EthdevCreate creates a ethdev and stores it in the portmngr ethdev store
func (pm *PortMngr) EthdevCreate(name string, params *ethdev.LinkParams) (*ethdev.Ethdev, error) {
	var e ethdev.Ethdev
	if pm.EthdevStore.Contains(name) {
		return nil, errors.New("ethdev with this name exists")
	}

	// initialize
	if err := e.Init(name, params, func() {
		pm.EthdevStore.Delete(name)
	}); err != nil {
		return nil, err
	}

	// add node to list
	pm.EthdevStore.Set(name, &e)
	log.Infof("ethdev %s created", name)
	return &e, nil
}

// TODO Extend LinkUpDown for ALL other port types
// LinkUpDown sets the given link (depicted with name) to up or down
func (pm *PortMngr) LinkUpDown(name string, status bool) error {
	ethdev := pm.EthdevStore.Get(name)
	if ethdev == nil {
		return errors.New("port doesn't exists")
	}

	if status {
		return ethdev.SetLinkUp()
	}
	return ethdev.SetLinkDown()
}

func (pm *PortMngr) GetPortStatsString(name string) (string, error) {
	var result string
	err := pm.EthdevStore.Iterate(func(key string, ethdev *ethdev.Ethdev) error {
		result += fmt.Sprintf("  %s \n", ethdev.DevName())
		stats, err := ethdev.GetPortStatsString()
		result += fmt.Sprintf("%s \n", stats)
		result += "\n"
		return err
	})
	return result, err
}

func (pm *PortMngr) EthdevList(name string) (string, error) {
	result := ""
	err := pm.EthdevStore.Iterate(func(key string, ethdev *ethdev.Ethdev) error {
		result += fmt.Sprintf("  %s \n", ethdev.DevName())
		devInfo, err := ethdev.GetPortInfoString()
		result += fmt.Sprintf("%s \n", devInfo)
		result += "\n"
		return err
	})
	return result, err
}
