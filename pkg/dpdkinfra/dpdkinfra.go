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

	// create store and initialize PortMngr and PipeMngr
	log.Info("Create Pktmbuf store...")
	di.PktmbufStore = store.NewStore[*pktmbuf.Pktmbuf]()

	log.Info("Initialize PortMngr...")
	di.PortMngr = &portmngr.PortMngr{}
	di.PortMngr.Init()

	log.Info("Initialize PipeMngr...")
	di.PipeMngr = &pipemngr.PipeMngr{}
	di.PipeMngr.Init()

	log.Info("Dpdkinfra initialization ready!")

	return nil
}

// empty & remove stores and cleanup initialized managers
func (di *DpdkInfra) Cleanup() error {
	di.PipeMngr.Cleanup()
	di.PortMngr.Cleanup()
	di.PktmbufStore.Clear()
	return dpdkswx.Runtime.Stop()
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
