// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package dpdkinfra

import (
	"errors"
	"log"

	"github.com/stolsma/go-p4dpdk-vswitch/pkg/dpdkswx"
)

type DpdkInfra struct {
	args          []string
	numArgs       int
	mbufStore     PktmbufStore
	ethdevStore   EthdevStore
	ringStore     RingStore
	pipelineStore PipelineStore
	tapStore      TapStore
}

func (di *DpdkInfra) Init(dpdkArgs []string) error {
	di.args = dpdkArgs
	numArgs, status := dpdkswx.EalInit(dpdkArgs)
	di.numArgs = numArgs
	if status != nil {
		return status
	}

	status = dpdkswx.ThreadInit()
	if status != nil {
		return status
	}
	log.Println("ThreadInit ok!")

	status = dpdkswx.MainThreadInit()
	if status != nil {
		return status
	}
	log.Println("MainThreadInit ok!")

	// create stores
	di.mbufStore = CreatePktmbufStore()
	di.ethdevStore = CreateEthdevStore()
	di.ringStore = CreateRingStore()
	di.tapStore = CreateTapStore()
	di.pipelineStore = CreatePipelineStore()

	return nil
}

func (di *DpdkInfra) Cleanup() {
	/*
		// thread 1 pipeline PIPELINE0 disable
		num, err = dpdkinfra.PipelineDisable(1, "PIPELINE0")
		if num != 0 {
			log.Fatalln("PipelineDisable num:", num)
		}
		if err != nil {
			log.Fatalln("PipelineDisable err:", err)
		}
		log.Println("Pipeline Disabled!")
	*/

	// empty & remove stores
	di.pipelineStore.Clear()
	di.tapStore.Clear()
	di.mbufStore.Clear()

	// TODO: cleanup EAL memory etc...

	/*
		err = dpdkinfra.EalCleanup()
		if err != nil {
			log.Fatalln("EAL cleanup err:", err)
		}
		log.Println("EAL cleanup ready!")
	*/
}

func (di *DpdkInfra) MempoolCreate(name string, bufferSize uint32, poolSize uint32, cacheSize uint32, cpuId int) error {
	_, err := di.mbufStore.Create(name, bufferSize, poolSize, cacheSize, cpuId)
	return err
}

func (di *DpdkInfra) TapCreate(name string) error {
	_, err := di.tapStore.Create(name)
	return err
}

func (di *DpdkInfra) RingCreate(name string, size uint32, numaNode uint32) error {
	_, err := di.ringStore.Create(name, size, numaNode)
	return err
}

func (di *DpdkInfra) EthdevCreate(name string, params *dpdkswx.EthdevParams) error {
	_, err := di.ethdevStore.Create(name, params)
	return err
}

func (di *DpdkInfra) PipelineCreate(name string, numaNode int) error {
	_, err := di.pipelineStore.Create(name, numaNode)
	return err
}

func (di *DpdkInfra) PipelineAddInputPortTap(plName string, portId int, tName string, mName string, mtu int, bsz int) error {
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

	pipeline.PortIsValid()

	return pipeline.AddInputPortTap(portId, tap, pktmbuf, mtu, bsz)
}

func (di *DpdkInfra) PipelineAddOutputPortTap(plName string, portId int, tName string, bsz int) error {
	pipeline := di.pipelineStore.Find(plName)
	if pipeline == nil {
		return errors.New("pipeline doesn't exists")
	}

	tap := di.tapStore.Find(tName)
	if tap == nil {
		return errors.New("tap interface doesn't exists")
	}

	pipeline.PortIsValid()

	return pipeline.AddOutputPortTap(portId, tap, bsz)
}

func (di *DpdkInfra) PipelineBuild(name string, specfile string) error {
	pipeline := di.pipelineStore.Find(name)
	if pipeline == nil {
		return errors.New("pipeline doesn't exists")
	}

	return pipeline.Build(specfile)
}

func (di *DpdkInfra) PipelineCommit(name string) error {
	pipeline := di.pipelineStore.Find(name)
	if pipeline == nil {
		return errors.New("pipeline doesn't exists")
	}

	return pipeline.Commit()
}

func (di *DpdkInfra) PipelineEnable(name string, threadid uint32) error {
	pipeline := di.pipelineStore.Find(name)
	if pipeline == nil {
		return errors.New("pipeline doesn't exists")
	}

	return pipeline.Enable(threadid)
}

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

func (di *DpdkInfra) PipelineStats(name string) (string, error) {
	pipeline := di.pipelineStore.Find(name)
	if pipeline == nil {
		return "", errors.New("pipeline doesn't exists")
	}

	return pipeline.Stats(), nil
}

func Init(dpdkArgs []string) (*DpdkInfra, error) {
	var di DpdkInfra
	err := di.Init(dpdkArgs)
	if err != nil {
		return nil, err
	}
	return &di, nil
}
