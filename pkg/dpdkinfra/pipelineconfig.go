// SPDX-FileCopyrightText: 2022-present Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package dpdkinfra

import (
	"path"
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

// Create a pipeline with a given pipeline configuration
func (dpdki *DpdkInfra) CreatePipelineWithConfig(pipelineConfig *PipelineConfig) {
	// Create pipeline
	pipeName := pipelineConfig.GetName()
	pl, err := dpdki.PipelineCreate(pipeName, pipelineConfig.GetNumaNode())
	if err != nil {
		log.Fatalf("%s create err: %d", pipeName, err)
	}
	log.Infof("%s created!", pipeName)

	// Add input ports to pipeline
	for i, t := range pipelineConfig.InputPorts {
		pName := t.GetIfaceName()
		port := dpdki.GetPort(pName)
		if port == nil {
			log.Fatalf("Pipeconfig %s input device %s does not exist", pipeName, pName)
		}

		err = port.BindToPipelineInputPort(pl, i, t.GetRxQueue(), t.GetBsz())
		if err != nil {
			log.Fatalf("AddInPort %s:%s err: %d", pipeName, pName, err)
		}

		log.Infof("AddInPort %s:%s ready!", pipeName, pName)
	}

	// Add output ports to pipeline
	for i, t := range pipelineConfig.OutputPorts {
		pName := t.GetIfaceName()
		port := dpdki.GetPort(pName)
		if port == nil {
			log.Fatalf("Pipeconfig %s input device %s does not exist", pipeName, pName)
		}

		err = port.BindToPipelineOutputPort(pl, i, t.GetTxQueue(), t.GetBsz())
		if err != nil {
			log.Fatalf("AddOutPort %s:%s err: %d", pipeName, pName, err)
		}

		log.Infof("AddOutPort %s:%s ready!", pipeName, pName)
	}

	// Build the pipeline program
	err = dpdki.PipelineBuild(pipeName, pipelineConfig.GetSpec())
	if err != nil {
		log.Fatalf("Pipelinebuild %s specfile: %s err: %d", pipeName, pipelineConfig.GetSpec(), err)
	}
	log.Infof("Pipeline %s Build!", pipeName)

	// Commit program to pipeline
	err = dpdki.PipelineCommit(pipeName)
	if err != nil {
		log.Fatalf("Pipelinecommit %s err: %d", pipeName, err)
	}
	log.Infof("Pipeline %s commited!", pipeName)

	// And run pipeline
	err = dpdki.PipelineEnable(pipeName, pipelineConfig.GetThreadID())
	if err != nil {
		log.Fatalf("PipelineEnable %s err: %d", pipeName, err)
	}
	log.Infof("Pipeline %s enabled!", pipeName)

	// Add Table startconfig
	for _, table := range pipelineConfig.Start.Tables {
		for _, line := range table.Data {
			dpdki.TableEntryAdd(pipeName, table.Name, line)
		}
	}

	// Commit Table changes to pipeline
	err = dpdki.PipelineCommit(pipeName)
	if err != nil {
		log.Fatalf("Table commit %s err: %d", pipeName, err)
	}
	log.Infof("Table config on pipeline %s commited!", pipeName)
}
