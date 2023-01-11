// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package portmngr

import (
	"errors"
	"fmt"

	"github.com/stolsma/go-p4pack/pkg/dpdkinfra/store"
	"github.com/stolsma/go-p4pack/pkg/dpdkswx/device"
	"github.com/stolsma/go-p4pack/pkg/dpdkswx/eal"
	"github.com/stolsma/go-p4pack/pkg/dpdkswx/ethdev"
	"github.com/stolsma/go-p4pack/pkg/dpdkswx/ring"
	"github.com/stolsma/go-p4pack/pkg/dpdkswx/sourcesink"
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
	SourceStore *store.Store[*sourcesink.Source]
	SinkStore   *store.Store[*sourcesink.Sink]
}

// Initialize the non system intrusive portmngr singleton parts
func (pm *PortMngr) Init() error {
	// create stores
	pm.EthdevStore = store.NewStore[*ethdev.Ethdev]()
	pm.RingStore = store.NewStore[*ring.Ring]()
	pm.TapStore = store.NewStore[*tap.Tap]()
	pm.SourceStore = store.NewStore[*sourcesink.Source]()
	pm.SinkStore = store.NewStore[*sourcesink.Sink]()

	return nil
}

func (pm *PortMngr) Cleanup() {
	// empty & remove stores
	pm.SinkStore.Clear()
	pm.SourceStore.Clear()
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

	if port := pm.SinkStore.Get(name); port != nil {
		return port
	}

	if port := pm.SourceStore.Get(name); port != nil {
		return port
	}

	return nil
}

func (pm *PortMngr) ContainsPort(name string) bool {
	return pm.TapStore.Contains(name) ||
		pm.EthdevStore.Contains(name) ||
		pm.RingStore.Contains(name) ||
		pm.SinkStore.Contains(name) ||
		pm.SourceStore.Contains(name)
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

	// iterate sink store
	if err := pm.SinkStore.Iterate(func(k string, v *sourcesink.Sink) error {
		return fn(k, v)
	}); err != nil {
		return err
	}

	// iterate source store
	if err := pm.SourceStore.Iterate(func(k string, v *sourcesink.Source) error {
		return fn(k, v)
	}); err != nil {
		return err
	}

	return nil
}

func (pm *PortMngr) SourceCreate(name string, params *sourcesink.SourceParams) (*sourcesink.Source, error) {
	var t sourcesink.Source
	if pm.ContainsPort(name) {
		return nil, errors.New("port with this name exists already")
	}

	if err := t.Init(name, params, func() {
		pm.SourceStore.Delete(name)
	}); err != nil {
		return nil, err
	}

	// add node to list
	pm.SourceStore.Set(name, &t)
	return &t, nil
}

func (pm *PortMngr) SinkCreate(name string, params *sourcesink.SinkParams) (*sourcesink.Sink, error) {
	var t sourcesink.Sink
	if pm.ContainsPort(name) {
		return nil, errors.New("port with this name exists already")
	}

	if err := t.Init(name, params, func() {
		pm.SinkStore.Delete(name)
	}); err != nil {
		return nil, err
	}

	// add node to list
	pm.SinkStore.Set(name, &t)
	return &t, nil
}

func (pm *PortMngr) TapCreate(name string, params *tap.Params) (*tap.Tap, error) {
	var t tap.Tap
	if pm.ContainsPort(name) {
		return nil, errors.New("port with this name exists already")
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

// RingCreate creates a ring and stores it in the portmngr ring store
func (pm *PortMngr) RingCreate(name string, params *ring.Params) (*ring.Ring, error) {
	var r ring.Ring
	if pm.ContainsPort(name) {
		return nil, errors.New("port with this name exists already")
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

// Attach (hotplug) the DPDK ethdev device defined by given DPDK device argument string
func (pm *PortMngr) AttachDevice(device string) (*eal.DevArgs, error) {
	var devArgs eal.DevArgs

	err := devArgs.Parse(device)
	if err != nil {
		return nil, fmt.Errorf("error parsing device argument string: %v", err)
	}

	err = eal.HotplugAdd(&devArgs)
	if err != nil {
		return nil, err
	}

	return &devArgs, nil
}

// Detach (hotplug) the DPDK ethdev device defined by given DPDK device argument string
func (pm *PortMngr) DetachDevice(device string) (*eal.DevArgs, error) {
	var devArgs eal.DevArgs

	err := devArgs.Parse(device)
	if err != nil {
		return nil, fmt.Errorf("error parsing device argument string: %v", err)
	}

	// first check if all ports related to the requested device are free
	err = pm.EthdevStore.Iterate(func(k string, v *ethdev.Ethdev) error {
		if v.DevName() == devArgs.Name() {
			return errors.New("some ports on the device are still used")
		}
		return nil
	})
	if err != nil {
		return &devArgs, err
	}

	// then detach device
	err = eal.HotplugRemove(&devArgs)
	if err != nil {
		return nil, err
	}

	return &devArgs, nil
}

// Get all raw DPDK ethdev ports
func (pm *PortMngr) GetAttachedEthdevPorts() ([]*ethdev.Ethdev, error) {
	return ethdev.GetAttachedPorts()
}

// Get all unused raw DPDK ethdev ports
func (pm *PortMngr) GetUnusedEthdevPorts() ([]*ethdev.Ethdev, error) {
	var ports []*ethdev.Ethdev

	rawPorts, err := ethdev.GetAttachedPorts()
	if err != nil {
		return nil, err
	}

	for _, rp := range rawPorts {
		if err := pm.EthdevStore.Iterate(func(k string, v *ethdev.Ethdev) error {
			if v.SamePort(rp) {
				return errors.New("port is used")
			}
			return nil
		}); err != nil {
			continue
		}
		ports = append(ports, rp)
	}

	return ports, nil
}

// EthdevCreate creates a ethdev and stores it in the portmngr ethdev store
func (pm *PortMngr) EthdevCreate(name string, params *ethdev.Params) (*ethdev.Ethdev, error) {
	var e ethdev.Ethdev
	if pm.ContainsPort(name) {
		return nil, errors.New("port with this name exists already")
	}

	// initialize struct, then initialize port
	e.Init(name)
	if err := e.Initialize(params, func() {
		pm.EthdevStore.Delete(name)
	}); err != nil {
		return nil, err
	}

	// add node to list
	pm.EthdevStore.Set(name, &e)
	log.Infof("ethdev %s created", name)
	return &e, nil
}

type EthdevPortFilter uint

const (
	AllEthdevPorts     EthdevPortFilter = iota + 1 // All created ethdev ports
	UnboundEthdevPorts                             // All ethdev ports created but not bound with one of its queues to a pipeline
	BoundEthdevPorts                               // All ethdev ports created and with one or more queues bound to a pipeline
)

// Get DPDK ethdev ports with filter applied
func (pm *PortMngr) GetEthdevPorts(filter EthdevPortFilter) ([]*ethdev.Ethdev, error) {
	var ports []*ethdev.Ethdev

	pm.EthdevStore.Iterate(func(k string, e *ethdev.Ethdev) error {
		switch filter {
		case UnboundEthdevPorts:
			if e.IsBound() {
				return nil
			}
		case BoundEthdevPorts:
			if !e.IsBound() {
				return nil
			}
		}
		ports = append(ports, e)
		return nil
	})

	return ports, nil
}

// LinkUp sets the given link (depicted with port name) to up if supported
// TODO remove, instead get port and do directly
func (pm *PortMngr) LinkUp(name string) error {
	port := pm.GetPort(name)
	if port == nil {
		return errors.New("port doesn't exists")
	}

	return port.SetLinkUp()
}

// LinkDown sets the given link (depicted with port name) to down if supported
// TODO remove, instead get port and do directly
func (pm *PortMngr) LinkDown(name string) error {
	port := pm.GetPort(name)
	if port == nil {
		return errors.New("port doesn't exists")
	}

	return port.SetLinkDown()
}

// returns the port info array of the requested port or all ports if no name given
func (pm *PortMngr) GetPortInfo(name string) (map[string]map[string]map[string]string, error) {
	result := make(map[string]map[string]map[string]string)
	var err error

	makeInfo := func(key string, port PortType) error {
		result[key] = make(map[string]map[string]string)
		result[key]["header"] = make(map[string]string)
		result[key]["header"]["name"] = port.Name()
		result[key]["header"]["type"] = port.Type()

		rxPipe := ""
		if err := port.IterateRxQueues(func(i uint16, q device.Queue) error {
			rxPipe += fmt.Sprintf("%s:%d ", q.Pipeline(), q.PipelinePort())
			return nil
		}); err != nil {
			return err
		}
		result[key]["header"]["rxqueuebound"] = rxPipe

		txPipe := ""
		if err := port.IterateTxQueues(func(i uint16, q device.Queue) error {
			txPipe += fmt.Sprintf("%s:%d ", q.Pipeline(), q.PipelinePort())
			return nil
		}); err != nil {
			return err
		}
		result[key]["header"]["txqueuebound"] = txPipe

		// TODO add linkstate!

		info, err := port.GetPortInfo()
		if err == device.ErrNotImplemented {
			result[key]["err"] = make(map[string]string)
			result[key]["err"]["err"] = fmt.Sprintf("%v", err)
			return nil
		}

		// merge
		result[key]["info"] = make(map[string]string)
		for k, v := range info {
			result[key]["info"][k] = v
		}

		return err
	}

	if name != "" {
		port := pm.GetPort(name)
		if port == nil {
			return result, fmt.Errorf("port with name %v not found", name)
		}
		err = makeInfo(name, port)
	} else {
		err = pm.IteratePorts(func(key string, port PortType) error {
			return makeInfo(key, port)
		})
	}

	return result, err
}

// returns the port statistics string of the requested port or all ports if no name given
func (pm *PortMngr) GetPortStats(name string) (map[string]map[string]map[string]string, error) {
	result := make(map[string]map[string]map[string]string)
	var err error

	makeStats := func(key string, port PortType) error {
		result[key] = make(map[string]map[string]string)
		result[key]["header"] = make(map[string]string)
		result[key]["header"]["name"] = port.Name()
		result[key]["header"]["type"] = port.Type()

		rxPipe := ""
		if err := port.IterateRxQueues(func(i uint16, q device.Queue) error {
			rxPipe += fmt.Sprintf("%s:%d ", q.Pipeline(), q.PipelinePort())
			return nil
		}); err != nil {
			return err
		}
		result[key]["header"]["rxqueuebound"] = rxPipe

		txPipe := ""
		if err := port.IterateTxQueues(func(i uint16, q device.Queue) error {
			txPipe += fmt.Sprintf("%s:%d ", q.Pipeline(), q.PipelinePort())
			return nil
		}); err != nil {
			return err
		}
		result[key]["header"]["txqueuebound"] = txPipe

		// TODO add linkstate!

		stats, err := port.GetPortStats()
		if err == device.ErrNotImplemented {
			result[key]["err"] = make(map[string]string)
			result[key]["err"]["err"] = fmt.Sprintf("%v", err)
			return nil
		}

		// merge
		result[key]["stats"] = make(map[string]string)
		for k, v := range stats {
			result[key]["stats"][k] = v
		}

		return err
	}

	if name != "" {
		port := pm.GetPort(name)
		if port == nil {
			return result, fmt.Errorf("port with name %v not found", name)
		}
		err = makeStats(name, port)
	} else {
		err = pm.IteratePorts(func(key string, port PortType) error {
			return makeStats(key, port)
		})
	}

	return result, err
}
