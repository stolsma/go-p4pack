// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package dpdkinfra

import (
	"errors"
	"fmt"

	"github.com/stolsma/go-p4pack/pkg/dpdkswx"
	"github.com/stolsma/go-p4pack/pkg/dpdkswx/ethdev"
	"github.com/stolsma/go-p4pack/pkg/dpdkswx/pipeline"
	"github.com/stolsma/go-p4pack/pkg/dpdkswx/swxruntime"
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
	args          []string
	numArgs       int
	mbufStore     PktmbufStore
	ethdevStore   EthdevStore
	ringStore     RingStore
	tapStore      TapStore
	pipelineStore PipelineStore
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

	// create dpdkinfra singleton
	dpdki = &DpdkInfra{}

	// initialize the dpdkswx runtime
	dpdki.args = dpdkArgs
	nArgs, err := dpdkswx.Runtime.Start(dpdkArgs)
	dpdki.numArgs = nArgs
	if err != nil {
		return nil, err
	}

	// initialize the dpdki singleton with non system intrusive parts
	err = dpdki.Init()
	if err != nil {
		return nil, err
	}

	return dpdki, nil
}

// Initialize the non system intrusive dpdkinfra singleton parts (i.e. excluding the dpdkswx runtime parts!)
func (di *DpdkInfra) Init() error {
	// create stores
	di.mbufStore = CreatePktmbufStore()
	di.ethdevStore = CreateEthdevStore()
	di.ringStore = CreateRingStore()
	di.tapStore = CreateTapStore()
	di.pipelineStore = CreatePipelineStore()

	return nil
}

func (di *DpdkInfra) Cleanup() {
	// TODO remove all pipelines from swx runtime singleton and stop the dpdkinfra singleton subparts!

	// empty & remove stores
	di.pipelineStore.Clear()
	di.tapStore.Clear()
	di.ringStore.Clear()
	di.ethdevStore.Clear()
	di.mbufStore.Clear()
}

func (di *DpdkInfra) PktMbufCreate(name string, bufferSize uint, poolSize uint32, cacheSize uint32, cpuID int) (err error) {
	_, err = di.mbufStore.Create(name, bufferSize, poolSize, cacheSize, cpuID)
	return err
}

func (di *DpdkInfra) TapCreate(name string, tapConfig *TapConfig) error {
	_, err := di.tapStore.Create(name, tapConfig)
	return err
}

func (di *DpdkInfra) TapList(name string) (string, error) {
	result := ""
	err := di.tapStore.Iterate(func(key string, tap *Tap) error {
		result += fmt.Sprintf("  %s \n", tap.Name())
		return nil
	})
	return result, err
}

func (di *DpdkInfra) RingCreate(name string, size uint, numaNode uint32) error {
	_, err := di.ringStore.Create(name, size, numaNode)
	return err
}

func (di *DpdkInfra) EthdevCreate(name string, params *ethdev.LinkParams) error {
	_, err := di.ethdevStore.Create(name, params)
	return err
}

func (di *DpdkInfra) LinkUpDown(name string, status bool) error {
	ethdev := di.ethdevStore.Find(name)
	if ethdev == nil {
		return errors.New("interface doesn't exists")
	}

	if status {
		return ethdev.SetLinkUp()
	}
	return ethdev.SetLinkDown()
}

func (di *DpdkInfra) GetPortStatsString(name string) (string, error) {
	var result string
	err := di.ethdevStore.Iterate(func(key string, ethdev *ethdev.Ethdev) error {
		result += fmt.Sprintf("  %s \n", ethdev.DevName())
		stats, err := ethdev.GetPortStatsString()
		result += fmt.Sprintf("%s \n", stats)
		result += "\n"
		return err
	})
	return result, err
}

func (di *DpdkInfra) EthdevList(name string) (string, error) {
	result := ""
	err := di.ethdevStore.Iterate(func(key string, ethdev *ethdev.Ethdev) error {
		result += fmt.Sprintf("  %s \n", ethdev.DevName())
		devInfo, err := ethdev.GetPortInfoString()
		result += fmt.Sprintf("%s \n", devInfo)
		result += "\n"
		return err
	})
	return result, err
}

func (di *DpdkInfra) PipelineCreate(plName string, numaNode int) (err error) {
	err = dpdkswx.Runtime.ExecOnMain(func(*swxruntime.MainCtx) {
		_, err = di.pipelineStore.Create(plName, numaNode)
	})
	return
}

func (di *DpdkInfra) PipelineAddInputPort(plName string, portID int, pName string, mName string, mtu int, rxQueue int, bsz int) error {
	if di.tapStore.Find(pName) != nil {
		return di.PipelineAddInputPortTap(plName, portID, pName, mName, mtu, bsz)
	}
	if di.ethdevStore.Find(pName) != nil {
		return di.PipelineAddInputPortEthDev(plName, portID, pName, rxQueue, bsz)
	}
	return errors.New("interface doesn't exists")
}

func (di *DpdkInfra) PipelineAddInputPortTap(plName string, portID int, tName string, mName string, mtu int, bsz int) error {
	pipeline := di.pipelineStore.Find(plName)
	if pipeline == nil {
		return errors.New("pipeline doesn't exists")
	}

	tap := di.tapStore.Find(tName)
	if tap == nil {
		return errors.New("tap interface doesn't exists")
	}

	pktmbuf := di.mbufStore.Find(mName)
	if pktmbuf == nil {
		return errors.New("mempool doesn't exists")
	}

	// TODO use this call
	pipeline.PortIsValid()

	return pipeline.AddInputPortTap(portID, int(tap.Fd()), pktmbuf.Mempool(), mtu, bsz)
}

func (di *DpdkInfra) PipelineAddInputPortEthDev(plName string, portID int, tName string, rxQueue int, bsz int) error {
	pipeline := di.pipelineStore.Find(plName)
	if pipeline == nil {
		return errors.New("pipeline doesn't exists")
	}

	ethDev := di.ethdevStore.Find(tName)
	if ethDev == nil {
		return errors.New("ethdev interface doesn't exists")
	}

	// TODO use this call
	pipeline.PortIsValid()

	return pipeline.AddInputPortEthDev(portID, ethDev.DevName(), rxQueue, bsz)
}

func (di *DpdkInfra) PipelineAddOutputPort(plName string, portID int, pName string, txQueue int, bsz int) error {
	if di.tapStore.Find(pName) != nil {
		return di.PipelineAddOutputPortTap(plName, portID, pName, bsz)
	}
	if di.ethdevStore.Find(pName) != nil {
		return di.PipelineAddOutputPortEthDev(plName, portID, pName, txQueue, bsz)
	}
	return errors.New("interface doesn't exists")
}

func (di *DpdkInfra) PipelineAddOutputPortTap(plName string, portID int, tName string, bsz int) error {
	pipeline := di.pipelineStore.Find(plName)
	if pipeline == nil {
		return errors.New("pipeline doesn't exists")
	}

	tap := di.tapStore.Find(tName)
	if tap == nil {
		return errors.New("tap interface doesn't exists")
	}

	// TODO use this call
	pipeline.PortIsValid()

	return pipeline.AddOutputPortTap(portID, int(tap.Fd()), bsz)
}

func (di *DpdkInfra) PipelineAddOutputPortEthDev(plName string, portID int, tName string, txQueue int, bsz int) error {
	pipeline := di.pipelineStore.Find(plName)
	if pipeline == nil {
		return errors.New("pipeline doesn't exists")
	}

	ethDev := di.ethdevStore.Find(tName)
	if ethDev == nil {
		return errors.New("ethdev interface doesn't exists")
	}

	// TODO use this call
	pipeline.PortIsValid()

	return pipeline.AddOutputPortEthdev(portID, ethDev.DevName(), txQueue, bsz)
}

func (di *DpdkInfra) PipelineBuild(plName string, specfile string) error {
	pipeline := di.pipelineStore.Find(plName)
	if pipeline == nil {
		return errors.New("pipeline doesn't exists")
	}

	return pipeline.BuildFromSpec(specfile)
}

func (di *DpdkInfra) PipelineCommit(plName string) error {
	pl := di.pipelineStore.Find(plName)
	if pl == nil {
		return errors.New("pipeline doesn't exists")
	}

	return pl.Commit(pipeline.CommitAbortOnFail)
}

func (di *DpdkInfra) PipelineEnable(plName string, threadID uint) error {
	pl := di.pipelineStore.Find(plName)
	return dpdkswx.Runtime.EnablePipeline(pl, threadID)
}

func (di *DpdkInfra) TableEntryAdd(plName string, tableName string, line string) error {
	pipeline := di.pipelineStore.Find(plName)
	if pipeline == nil {
		return errors.New("pipeline doesn't exists")
	}

	tableEntry := pipeline.TableEntryRead(tableName, line)
	if tableEntry == nil {
		return nil
	}

	err := pipeline.TableEntryAdd(tableName, tableEntry)
	return err
}

/*
func (di *DpdkInfra) PrintThreadStatus() {
	var i uint
	for i = 0; i < 16; i++ {
		if dpdkswx.ThreadIsRunning(i) {
			log.Printf("Thread %d running!", i)
		} else {
			log.Printf("Thread %d not running!", i)
		}
	}
}
*/

// get pipeline info. If plName is filled then that specific pipeline info is retrieved else the info
// of all pipelines is retrieved
func (di *DpdkInfra) PipelineInfo(plName string) (string, error) {
	if plName != "" {
		pl := di.pipelineStore.Find(plName)
		if pl == nil {
			return "", errors.New("pipeline doesn't exists")
		}
		return pl.Info(), nil
	}

	result := ""
	err := di.pipelineStore.Iterate(func(key string, pl *pipeline.Pipeline) error {
		result += fmt.Sprintf("%s: \n", pl.GetName())
		result += pl.Info()
		return nil
	})

	return result, err
}

// get pipeline statistics. If plName is filled then that specific pipeline statistics is retrieved else the statistics
// of all pipelines is retrieved
func (di *DpdkInfra) PipelineStats(plName string) (string, error) {
	if plName != "" {
		pipeline := di.pipelineStore.Find(plName)
		if pipeline == nil {
			return "", errors.New("pipeline doesn't exists")
		}
		return pipeline.Stats(), nil
	}

	result := ""
	err := di.pipelineStore.Iterate(func(key string, pl *pipeline.Pipeline) error {
		result += fmt.Sprintf("%s: \n", pl.GetName())
		result += pl.Stats()
		return nil
	})

	return result, err
}
