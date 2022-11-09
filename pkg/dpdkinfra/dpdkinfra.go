// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package dpdkinfra

import (
	"errors"
	"fmt"
	"log"

	"github.com/stolsma/go-p4pack/pkg/dpdkswx"
)

type DpdkInfra struct {
	args          []string
	numArgs       int
	mbufStore     PktmbufStore
	ethdevStore   EthdevStore
	ringStore     RingStore
	tapStore      TapStore
	pipelineStore PipelineStore
}

func Create() *DpdkInfra {
	var di DpdkInfra
	return &di
}

func CreateAndInit(dpdkArgs []string) (*DpdkInfra, error) {
	di := Create()
	err := di.Init(dpdkArgs)
	if err != nil {
		return nil, err
	}
	return di, nil
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

func (di *DpdkInfra) MempoolCreate(name string, bufferSize uint32, poolSize uint32, cacheSize uint32, cpuID int) error {
	_, err := di.mbufStore.Create(name, bufferSize, poolSize, cacheSize, cpuID)
	return err
}

func (di *DpdkInfra) TapCreate(name string) error {
	_, err := di.tapStore.Create(name)
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

func (di *DpdkInfra) RingCreate(name string, size uint32, numaNode uint32) error {
	_, err := di.ringStore.Create(name, size, numaNode)
	return err
}

func (di *DpdkInfra) EthdevCreate(name string, params *dpdkswx.EthdevParams) error {
	_, err := di.ethdevStore.Create(name, params)
	return err
}

func (di *DpdkInfra) PipelineCreate(plName string, numaNode int) error {
	_, err := di.pipelineStore.Create(plName, numaNode)
	return err
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

	pipeline.PortIsValid()

	return pipeline.AddInputPortTap(portID, int(tap.Fd()), pktmbuf, mtu, bsz)
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

	pipeline.PortIsValid()

	return pipeline.AddOutputPortTap(portID, int(tap.Fd()), bsz)
}

func (di *DpdkInfra) PipelineBuild(plName string, specfile string) error {
	pipeline := di.pipelineStore.Find(plName)
	if pipeline == nil {
		return errors.New("pipeline doesn't exists")
	}

	return pipeline.BuildFromSpec(specfile)
}

func (di *DpdkInfra) PipelineCommit(plName string) error {
	pipeline := di.pipelineStore.Find(plName)
	if pipeline == nil {
		return errors.New("pipeline doesn't exists")
	}

	return pipeline.Commit(dpdkswx.CommitAbortOnFail)
}

func (di *DpdkInfra) PipelineEnable(plName string, threadID uint32) error {
	pipeline := di.pipelineStore.Find(plName)
	if pipeline == nil {
		return errors.New("pipeline doesn't exists")
	}

	return pipeline.Enable(threadID)
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

// get pipeline info. If plName is filled then that specific pipeline info is retrieved else the info
// of all pipelines is retrieved
func (di *DpdkInfra) PipelineInfo(plName string) (string, error) {
	if plName != "" {
		pipeline := di.pipelineStore.Find(plName)
		if pipeline == nil {
			return "", errors.New("pipeline doesn't exists")
		}
		return pipeline.Info(), nil
	}

	result := ""
	err := di.pipelineStore.Iterate(func(key string, pipeline *dpdkswx.Pipeline) error {
		result += fmt.Sprintf("%s: \n", pipeline.GetName())
		result += pipeline.Info()
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
	err := di.pipelineStore.Iterate(func(key string, pipeline *dpdkswx.Pipeline) error {
		result += fmt.Sprintf("%s: \n", pipeline.GetName())
		result += pipeline.Stats()
		return nil
	})

	return result, err
}
