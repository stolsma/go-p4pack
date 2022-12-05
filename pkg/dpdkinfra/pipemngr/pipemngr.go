// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package pipemngr

import (
	"errors"
	"fmt"

	"github.com/stolsma/go-p4pack/pkg/dpdkinfra/store"
	"github.com/stolsma/go-p4pack/pkg/dpdkswx"
	"github.com/stolsma/go-p4pack/pkg/dpdkswx/pipeline"
	"github.com/stolsma/go-p4pack/pkg/dpdkswx/swxruntime"
	"github.com/stolsma/go-p4pack/pkg/logging"
	"github.com/yerden/go-dpdk/mempool"
)

var log logging.Logger

func init() {
	// keep the logger up to date, also after new log config
	logging.Register("dpdkinfra/pipemngr", func(logger logging.Logger) {
		log = logger
	})
}

type PipeMngr struct {
	PipelineStore *store.Store[*pipeline.Pipeline]
}

// Initialize the non system intrusive dpdkinfra singleton parts (i.e. excluding the dpdkswx runtime parts!)
func (pm *PipeMngr) Init() error {
	// create stores
	pm.PipelineStore = store.NewStore[*pipeline.Pipeline]()

	return nil
}

func (pm *PipeMngr) Cleanup() {
	// empty pipelinestore
	pm.PipelineStore.Clear()
}

func (pm *PipeMngr) PipelineCreate(plName string, numaNode int) (*pipeline.Pipeline, error) {
	var pl pipeline.Pipeline
	var innerErr error

	// execute pipeline creation on dpdk main thread
	err := dpdkswx.Runtime.ExecOnMain(func(*swxruntime.MainCtx) {
		// initialize pipeline record
		innerErr = pl.Init(plName, numaNode, func() {
			log.Infof("Remove pipeline %s from store", plName)
			pm.PipelineStore.Delete(plName)
		})
	})

	// check if something went wrong
	if err != nil {
		return nil, err
	} else if innerErr != nil {
		return nil, innerErr
	}

	// add node to list
	pm.PipelineStore.Set(plName, &pl)
	return &pl, nil
}

func (pm *PipeMngr) PipelineAddInputPortTap(plName string, portID int, tap int, mp *mempool.Mempool, mtu int, bsz int) error {
	pipeline := pm.PipelineStore.Get(plName)
	if pipeline == nil {
		return errors.New("pipeline doesn't exists")
	}

	return pipeline.AddInputPortTap(portID, tap, mp, mtu, bsz)
}

func (pm *PipeMngr) PipelineAddInputPortEthDev(plName string, portID int, devName string, rxQueue int, bsz int) error {
	pipeline := pm.PipelineStore.Get(plName)
	if pipeline == nil {
		return errors.New("pipeline doesn't exists")
	}

	return pipeline.AddInputPortEthDev(portID, devName, rxQueue, bsz)
}

func (pm *PipeMngr) PipelineAddOutputPortTap(plName string, portID int, tap int, bsz int) error {
	pipeline := pm.PipelineStore.Get(plName)
	if pipeline == nil {
		return errors.New("pipeline doesn't exists")
	}

	return pipeline.AddOutputPortTap(portID, tap, bsz)
}

func (pm *PipeMngr) PipelineAddOutputPortEthDev(plName string, portID int, devname string, txQueue int, bsz int) error {
	pipeline := pm.PipelineStore.Get(plName)
	if pipeline == nil {
		return errors.New("pipeline doesn't exists")
	}

	return pipeline.AddOutputPortEthdev(portID, devname, txQueue, bsz)
}

func (pm *PipeMngr) PipelineBuild(plName string, specfile string) error {
	pipeline := pm.PipelineStore.Get(plName)
	if pipeline == nil {
		return errors.New("pipeline doesn't exists")
	}

	// Check for valid number of ports
	if !pipeline.PortIsValid() {
		return errors.New("number of receive ports in this pipeline is 0 or not a power of 2")
	}

	return pipeline.BuildFromSpec(specfile)
}

func (pm *PipeMngr) PipelineCommit(plName string) error {
	pl := pm.PipelineStore.Get(plName)
	if pl == nil {
		return errors.New("pipeline doesn't exists")
	}

	return pl.Commit(pipeline.CommitAbortOnFail)
}

func (pm *PipeMngr) PipelineEnable(plName string, threadID uint) error {
	pl := pm.PipelineStore.Get(plName)
	return dpdkswx.Runtime.EnablePipeline(pl, threadID)
}

func (pm *PipeMngr) TableEntryAdd(plName string, tableName string, line string) error {
	pipeline := pm.PipelineStore.Get(plName)
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

// get pipeline info. If plName is filled then that specific pipeline info is retrieved else the info
// of all pipelines is retrieved
func (pm *PipeMngr) PipelineInfo(plName string) (string, error) {
	if plName != "" {
		pl := pm.PipelineStore.Get(plName)
		if pl == nil {
			return "", errors.New("pipeline doesn't exists")
		}
		return pl.Info(), nil
	}

	result := ""
	err := pm.PipelineStore.Iterate(func(key string, pl *pipeline.Pipeline) error {
		result += fmt.Sprintf("%s: \n", pl.GetName())
		result += pl.Info()
		return nil
	})

	return result, err
}

// get pipeline statistics. If plName is filled then that specific pipeline statistics is retrieved else the statistics
// of all pipelines is retrieved
func (pm *PipeMngr) PipelineStats(plName string) (string, error) {
	if plName != "" {
		pipeline := pm.PipelineStore.Get(plName)
		if pipeline == nil {
			return "", errors.New("pipeline doesn't exists")
		}
		return pipeline.Stats(), nil
	}

	result := ""
	err := pm.PipelineStore.Iterate(func(key string, pl *pipeline.Pipeline) error {
		result += fmt.Sprintf("%s: \n", pl.GetName())
		result += pl.Stats()
		return nil
	})

	return result, err
}
