// SPDX-FileCopyrightText: 2022-present Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"errors"
	"fmt"
	"path"

	"github.com/stolsma/go-p4pack/pkg/dpdkinfra"
)

type PipelineConfig struct {
	Name        string           `json:"name"`
	NumaNode    int              `json:"numanode"`
	BasePath    string           `json:"basepath"`
	Spec        string           `json:"spec"`
	ThreadID    uint             `json:"threadid"`
	OutputPorts []*OutPortConfig `json:"outputports"`
	InputPorts  []*InPortConfig  `json:"inputports"`
	Start       *StartConfig     `json:"start"`
}

func (pc *PipelineConfig) GetName() string {
	return pc.Name
}

func (pc *PipelineConfig) GetNumaNode() int {
	return pc.NumaNode
}

func (pc *PipelineConfig) SetBasePath(basePath string) bool {
	if pc.BasePath != "" {
		return false
	}
	pc.BasePath = basePath
	return true
}

func (pc *PipelineConfig) GetBasePath() string {
	return pc.BasePath
}

func (pc *PipelineConfig) GetSpec() string {
	return path.Join(pc.BasePath, pc.Spec)
}

func (pc *PipelineConfig) GetThreadID() uint {
	return pc.ThreadID
}

type InPortConfig struct {
	IfaceName string `json:"ifacename"`
	RxQueue   uint   `json:"rxqueue"`
	Bsz       uint   `json:"bsz"`
}

func (pc *InPortConfig) GetIfaceName() string {
	return pc.IfaceName
}

func (pc *InPortConfig) GetRxQueue() uint {
	return pc.RxQueue
}

func (pc *InPortConfig) GetBsz() uint {
	return pc.Bsz
}

type OutPortConfig struct {
	IfaceName string `json:"ifacename"`
	TxQueue   uint   `json:"txqueue"`
	Bsz       uint   `json:"bsz"`
}

func (pc *OutPortConfig) GetIfaceName() string {
	return pc.IfaceName
}

func (pc *OutPortConfig) GetTxQueue() uint {
	return pc.TxQueue
}

func (pc *OutPortConfig) GetBsz() uint {
	return pc.Bsz
}

type StartConfig struct {
	Tables []TableConfig `json:"tables"`
}

type TableConfig struct {
	Name string   `json:"name"`
	Data []string `json:"data"`
}

// Create pipelines with a given pipeline configuration list
func (c *Config) ApplyPipeline(basePath string) error {
	dpdki := dpdkinfra.Get()
	if dpdki == nil {
		return errors.New("dpdkinfra module is not initialized")
	}

	// Create pipeline
	for _, pConfig := range c.Pipelines {
		pConfig.SetBasePath(basePath)
		pipeName := pConfig.GetName()
		pl, err := dpdki.PipelineCreate(pipeName, pConfig.GetNumaNode())
		if err != nil {
			return fmt.Errorf("%s create err: %v", pipeName, err)
		}
		log.Infof("%s created!", pipeName)

		// Add input ports to pipeline
		for i, t := range pConfig.InputPorts {
			pName := t.GetIfaceName()
			port := dpdki.GetPort(pName)
			if port == nil {
				return fmt.Errorf("Pipeconfig %s input device %s does not exist", pipeName, pName)
			}

			err = port.BindToPipelineInputPort(pl, i, t.GetRxQueue(), t.GetBsz())
			if err != nil {
				return fmt.Errorf("AddInPort %s:%s err: %v", pipeName, pName, err)
			}

			log.Infof("AddInPort %s:%s ready!", pipeName, pName)
		}

		// Add output ports to pipeline
		for i, t := range pConfig.OutputPorts {
			pName := t.GetIfaceName()
			port := dpdki.GetPort(pName)
			if port == nil {
				return fmt.Errorf("pipeconfig %s input device %s does not exist", pipeName, pName)
			}

			err = port.BindToPipelineOutputPort(pl, i, t.GetTxQueue(), t.GetBsz())
			if err != nil {
				return fmt.Errorf("AddOutPort %s:%s err: %v", pipeName, pName, err)
			}

			log.Infof("AddOutPort %s:%s ready!", pipeName, pName)
		}

		// Build the pipeline program
		err = dpdki.PipelineBuild(pipeName, pConfig.GetSpec())
		if err != nil {
			return fmt.Errorf("pipelinebuild %s specfile: %s err: %v", pipeName, pConfig.GetSpec(), err)
		}
		log.Infof("Pipeline %s build with specfile: %s ", pipeName, pConfig.GetSpec())

		// Commit program to pipeline
		err = dpdki.PipelineCommit(pipeName)
		if err != nil {
			return fmt.Errorf("pipeline %s commit err: %v", pipeName, err)
		}
		log.Infof("Pipeline %s commited!", pipeName)

		// And run pipeline
		err = dpdki.PipelineEnable(pipeName, pConfig.GetThreadID())
		if err != nil {
			return fmt.Errorf("pipelineEnable %s err: %v", pipeName, err)
		}
		log.Infof("Pipeline %s enabled!", pipeName)

		// Add Table startconfig
		for _, table := range pConfig.Start.Tables {
			for _, line := range table.Data {
				err := dpdki.TableEntryAdd(pipeName, table.Name, line)
				if err != nil {
					return fmt.Errorf("table entry add went wrong (Pipeline: %s, Table: %s, Line: %s). err: %v",
						pipeName, table.Name, line, err)
				}
			}
		}

		// Commit Table changes to pipeline
		err = dpdki.PipelineCommit(pipeName)
		if err != nil {
			return fmt.Errorf("pipeline %s commit err: %v", pipeName, err)
		}
		log.Infof("Table config on pipeline %s commited!", pipeName)
	}

	return nil
}
