// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package portmngr

import (
	"errors"
	"fmt"

	"github.com/stolsma/go-p4pack/pkg/dpdkinfra/store"
	"github.com/stolsma/go-p4pack/pkg/dpdkswx/device"
	"github.com/stolsma/go-p4pack/pkg/dpdkswx/ethdev"
	"github.com/stolsma/go-p4pack/pkg/dpdkswx/ring"
	"github.com/stolsma/go-p4pack/pkg/dpdkswx/tap"
	"github.com/stolsma/go-p4pack/pkg/logging"
)

var log logging.Logger

func init() {
	// keep the logger up to date, also after new log config
	logging.Register("dpdkinfra/portmngr", func(logger logging.Logger) {
		log = logger
	})
}

type PortType interface {
	store.ValueInterface
	device.Type
}

type PortMngr struct {
	EthdevStore *store.Store[*ethdev.Ethdev]
	RingStore   *store.Store[*ring.Ring]
	TapStore    *store.Store[*tap.Tap]
}

// Initialize the non system intrusive portmngr singleton parts
func (pm *PortMngr) Init() error {
	// create stores
	pm.EthdevStore = store.NewStore[*ethdev.Ethdev]()
	pm.RingStore = store.NewStore[*ring.Ring]()
	pm.TapStore = store.NewStore[*tap.Tap]()

	return nil
}

func (pm *PortMngr) Cleanup() {
	// empty & remove stores
	pm.TapStore.Clear()
	pm.RingStore.Clear()
	pm.EthdevStore.Clear()
}

func (pm *PortMngr) GetPort(name string) PortType {
	if port := pm.TapStore.Get(name); port != nil {
		return port
	}

	if port := pm.EthdevStore.Get(name); port != nil {
		return port
	}

	if port := pm.RingStore.Get(name); port != nil {
		return port
	}

	return nil
}

// Iterate over the contents of all the device stores
func (pm *PortMngr) IteratePorts(fn func(key string, value PortType) error) error {
	// iterate tap store
	if err := pm.TapStore.Iterate(func(k string, v *tap.Tap) error {
		return fn(k, v)
	}); err != nil {
		return err
	}

	// iterate ethdev store
	if err := pm.EthdevStore.Iterate(func(k string, v *ethdev.Ethdev) error {
		return fn(k, v)
	}); err != nil {
		return err
	}

	// iterate ring store
	if err := pm.RingStore.Iterate(func(k string, v *ring.Ring) error {
		return fn(k, v)
	}); err != nil {
		return err
	}

	return nil
}

func (pm *PortMngr) TapCreate(name string, params *tap.Params) (*tap.Tap, error) {
	var t tap.Tap
	if pm.TapStore.Contains(name) {
		return nil, errors.New("tap with this name exists")
	}

	if err := t.Init(name, params, func() {
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
	err := pm.TapStore.Iterate(func(key string, tap *tap.Tap) error {
		result += fmt.Sprintf("  %s \n", tap.Name())
		return nil
	})
	return result, err
}

// RingCreate creates a ring and stores it in the portmngr ring store
func (pm *PortMngr) RingCreate(name string, params *ring.Params) (*ring.Ring, error) {
	var r ring.Ring
	if pm.RingStore.Contains(name) {
		return nil, errors.New("ring with this name exists")
	}

	// initialize
	if err := r.Init(name, params, func() {
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
func (pm *PortMngr) EthdevCreate(name string, params *ethdev.Params) (*ethdev.Ethdev, error) {
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

// LinkUp sets the given link (depicted with port name) to up if supported
func (pm *PortMngr) LinkUp(name string) error {
	port := pm.GetPort(name)
	if port == nil {
		return errors.New("port doesn't exists")
	}

	return port.SetLinkUp()
}

// LinkDown sets the given link (depicted with port name) to down if supported
func (pm *PortMngr) LinkDown(name string) error {
	port := pm.GetPort(name)
	if port == nil {
		return errors.New("port doesn't exists")
	}

	return port.SetLinkDown()
}

// returns the port statistics string of the requested port or all ports if no name given
func (pm *PortMngr) GetPortStatsString(name string) (map[string]string, error) {
	result := make(map[string]string)
	var err error

	makeString := func(key string, port PortType) error {
		result[key] += fmt.Sprintf("\n  %v <%v", port.Name(), port.Type())
		if port.PipelineInPort() != device.NotBound {
			result[key] += fmt.Sprintf(", %v:%v", port.PipelineIn(), port.PipelineInPort())
		}
		if port.PipelineOutPort() != device.NotBound {
			result[key] += fmt.Sprintf(", %v:%v", port.PipelineOut(), port.PipelineOutPort())
		}
		// TODO add linkstate!
		result[key] += ">\n"

		stats, err := port.GetPortStatsString()
		if err == device.ErrNotImplemented {
			result[key] += fmt.Sprintf("\tPort statistics: %v\n", err)
			return nil
		}
		result[key] += stats
		return err
	}

	if name != "" {
		port := pm.GetPort(name)
		if port == nil {
			return result, fmt.Errorf("port with name %v not found", name)
		}
		err = makeString(name, port)
	} else {
		err = pm.IteratePorts(func(key string, port PortType) error {
			return makeString(key, port)
		})
	}

	return result, err
}
