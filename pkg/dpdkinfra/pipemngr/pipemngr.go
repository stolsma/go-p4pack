// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package pipemngr

import (
	"errors"

	"github.com/stolsma/go-p4pack/pkg/dpdkinfra/store"
	"github.com/stolsma/go-p4pack/pkg/dpdkswx/pipeline"
	"github.com/stolsma/go-p4pack/pkg/logging"
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

	// initialize pipeline record
	if err := pl.Init(plName, numaNode, func() {
		log.Infof("Remove pipeline %s from store", plName)
		pm.PipelineStore.Delete(plName)
	}); err != nil {
		return nil, err
	}

	// add node to list
	pm.PipelineStore.Set(plName, &pl)
	return &pl, nil
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
	return pl.SetEnabled(threadID)
}

func (pm *PipeMngr) PipelineDisable(plName string) error {
	pl := pm.PipelineStore.Get(plName)
	return pl.SetDisabled()
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

var ErrPipelineInfoGet = errors.New("pipeline info couldn't be retrieved")

type PipelineInfoList map[string]*pipeline.Info

// get pipeline info list. If plName is filled then that specific pipeline info is retrieved else the info
// of all pipelines is retrieved
func (pm *PipeMngr) PipelineInfo(plName string) (PipelineInfoList, error) {
	result := make(PipelineInfoList, 0)

	if plName != "" {
		pl := pm.PipelineStore.Get(plName)
		if pl == nil {
			return result, errors.New("pipeline doesn't exists")
		}

		pli, err := pl.PipelineInfoGet()
		if err != nil {
			return result, ErrPipelineInfoGet
		}

		result[pl.GetName()] = pli

		return result, nil
	}

	err := pm.PipelineStore.Iterate(func(key string, pl *pipeline.Pipeline) error {
		pli, err := pl.PipelineInfoGet()
		if err != nil {
			return ErrPipelineInfoGet
		}

		result[pl.GetName()] = pli

		return nil
	})

	return result, err
}

type PipelineStatsList map[string]*pipeline.Stats

// get pipeline statistics. If plName is filled then that specific pipeline statistics is retrieved else the statistics
// of all pipelines is retrieved
func (pm *PipeMngr) PipelineStats(plName string) (PipelineStatsList, error) {
	result := make(PipelineStatsList)
	if plName != "" {
		pl := pm.PipelineStore.Get(plName)
		if pl == nil {
			return nil, errors.New("pipeline doesn't exists")
		}

		stats, err := pl.StatsRead()
		if err != nil {
			return nil, err
		}

		result[pl.GetName()] = stats
		return result, nil
	}

	err := pm.PipelineStore.Iterate(func(key string, pl *pipeline.Pipeline) error {
		stats, err := pl.StatsRead()
		if err != nil {
			return err
		}

		result[pl.GetName()] = stats
		return nil
	})

	return result, err
}
