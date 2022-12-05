// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package dpdkinfra

import (
	"errors"

	"github.com/stolsma/go-p4pack/pkg/dpdkinfra/pipemngr"
	"github.com/stolsma/go-p4pack/pkg/dpdkinfra/portmngr"
	"github.com/stolsma/go-p4pack/pkg/dpdkinfra/store"
	"github.com/stolsma/go-p4pack/pkg/dpdkswx"
	"github.com/stolsma/go-p4pack/pkg/dpdkswx/pktmbuf"
	"github.com/stolsma/go-p4pack/pkg/logging"
)

var dpdki *DpdkInfra
var log logging.Logger

func init() {
	// keep the logger up to date, also after new log config
	logging.Register("dpdkinfra", func(logger logging.Logger) {
		log = logger
	})
}

type DpdkInfra struct {
	args    []string
	numArgs int
	*portmngr.PortMngr
	*pipemngr.PipeMngr
	PktmbufStore *store.Store[*pktmbuf.Pktmbuf]
}

// return a pointer to the (initialized) dpdkinfra singleton
func Get() *DpdkInfra {
	return dpdki
}

// create and initialize the dpdkinfra singleton, return the current dpdkinfra singleton with error if it already exists
func CreateAndInit(dpdkArgs []string) (*DpdkInfra, error) {
	if dpdki != nil {
		return dpdki, errors.New("dpdki already initialized")
	}

	// create & initialize the dpdkinfra singleton
	dpdki = &DpdkInfra{}
	if err := dpdki.init(dpdkArgs); err != nil {
		return nil, err
	}

	return dpdki, nil
}

// Initialize the non system intrusive dpdkinfra singleton parts (i.e. excluding the dpdkswx runtime parts!)
func (di *DpdkInfra) init(dpdkArgs []string) error {
	// initialize the dpdkswx runtime
	dpdki.args = dpdkArgs
	nArgs, err := dpdkswx.Runtime.Start(dpdkArgs)
	dpdki.numArgs = nArgs
	if err != nil {
		return err
	}

	// create stores
	di.PktmbufStore = store.NewStore[*pktmbuf.Pktmbuf]()

	di.PortMngr = &portmngr.PortMngr{}
	di.PortMngr.Init()

	di.PipeMngr = &pipemngr.PipeMngr{}
	di.PipeMngr.Init()

	return nil
}

// empty & remove stores and cleanup initialized managers
func (di *DpdkInfra) Cleanup() {
	di.PipeMngr.Cleanup()
	di.PortMngr.Cleanup()
	di.PktmbufStore.Clear()
}

// PktmbufCreate creates a pktmuf and stores it in the dpdkinfra pktmbuf store
func (di *DpdkInfra) PktmbufCreate(
	name string, bufferSize uint, poolSize uint32, cacheSize uint32, cpuID int,
) (*pktmbuf.Pktmbuf, error) {
	var pm pktmbuf.Pktmbuf
	if di.PktmbufStore.Contains(name) {
		return nil, errors.New("pktmbuf mempool with this name exists")
	}

	// initialize
	if err := pm.Init(name, bufferSize, poolSize, cacheSize, cpuID, func() {
		di.PktmbufStore.Delete(name)
	}); err != nil {
		return nil, err
	}

	// add node to list
	di.PktmbufStore.Set(name, &pm)
	log.Infof("pktmbuf %s created", name)
	return &pm, nil
}

func (di *DpdkInfra) PipelineAddInputPort(plName string, portID int, portName string, mpName string, mtu int, rxQueue int, bsz int) error {
	if tap := di.TapStore.Get(portName); tap != nil {
		pktmbuf := di.PktmbufStore.Get(mpName)
		if pktmbuf == nil {
			return errors.New("mempool doesn't exists")
		}

		return di.PipelineAddInputPortTap(plName, portID, int(tap.Fd()), pktmbuf.Mempool(), mtu, bsz)
	}

	if ethdev := di.EthdevStore.Get(portName); ethdev != nil {
		return di.PipelineAddInputPortEthDev(plName, portID, ethdev.DevName(), rxQueue, bsz)
	}

	return errors.New("interface doesn't exists")
}

func (di *DpdkInfra) PipelineAddOutputPort(plName string, portID int, pName string, txQueue int, bsz int) error {
	if tap := di.TapStore.Get(pName); tap != nil {
		return di.PipelineAddOutputPortTap(plName, portID, int(tap.Fd()), bsz)
	}
	if ethdev := di.EthdevStore.Get(pName); ethdev != nil {
		return di.PipelineAddOutputPortEthDev(plName, portID, ethdev.DevName(), txQueue, bsz)
	}
	return errors.New("interface doesn't exists")
}
