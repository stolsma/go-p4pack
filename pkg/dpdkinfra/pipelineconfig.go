// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package dpdkinfra

import (
	"log"
	"path"
)

type PipelineConfig struct {
	Name        string
	NumaNode    int
	Spec        string
	ThreadID    uint32
	Mempools    []MempoolConfig
	OutputPorts []OutPortConfig
	InputPorts  []InPortConfig
	Start       StartConfig
}

func PathConfig(base string, pipeline *[]PipelineConfig) {
	pipes := *pipeline
	for i := range pipes {
		if pipes[i].Spec == "" {
			log.Fatalf("the configuration of Pipeline %s doesn't have Spec defined", pipes[i].GetName())
		}
		pipes[i].Spec = path.Join(base, pipes[i].Spec)
	}
}

func (pc *PipelineConfig) GetName() string {
	return pc.Name
}

func (pc *PipelineConfig) GetNumaNode() int {
	return pc.NumaNode
}

func (pc *PipelineConfig) GetSpec() string {
	return pc.Spec
}

func (pc *PipelineConfig) GetThreadID() uint32 {
	return pc.ThreadID
}

type MempoolConfig struct {
	Name       string
	BufferSize uint32
	PoolSize   uint32
	CacheSize  uint32
	CPUID      int
}

func (mpc *MempoolConfig) GetName() string {
	return mpc.Name
}

func (mpc *MempoolConfig) GetBufferSize() uint32 {
	return mpc.BufferSize
}

func (mpc *MempoolConfig) GetPoolSize() uint32 {
	return mpc.PoolSize
}

func (mpc *MempoolConfig) GetCacheSize() uint32 {
	return mpc.CacheSize
}

func (mpc *MempoolConfig) GetCPUID() int {
	return mpc.CPUID
}

type InPortConfig struct {
	IfaceName string
	Mempool   string
	MTU       int
	Bsz       int
}

func (pc *InPortConfig) GetIfaceName() string {
	return pc.IfaceName
}

func (pc *InPortConfig) GetMempool() string {
	return pc.Mempool
}

func (pc *InPortConfig) GetMTU() int {
	return pc.MTU
}

func (pc *InPortConfig) GetBsz() int {
	return pc.Bsz
}

type OutPortConfig struct {
	IfaceName string
	Bsz       int
}

func (pc *OutPortConfig) GetIfaceName() string {
	return pc.IfaceName
}

func (pc *OutPortConfig) GetBsz() int {
	return pc.Bsz
}

type StartConfig struct {
	Tables []TableConfig
}

type TableConfig struct {
	Name string
	Data []string
}

// Create pipelines through the DpdkInfra API
func (dpdki *DpdkInfra) PipelineWithConfig(pipelineConfig PipelineConfig) {
	// Create pipeline
	pipeName := pipelineConfig.GetName()
	err := dpdki.PipelineCreate(pipeName, pipelineConfig.GetNumaNode())
	if err != nil {
		log.Fatalf("%s create err: %d", pipeName, err)
	}
	log.Printf("%s created!", pipeName)

	// Create mempools
	// mempool MEMPOOL0 buffer 2304 pool 32K cache 256 cpu 0
	for _, m := range pipelineConfig.Mempools {
		name := m.GetName()
		err := dpdki.MempoolCreate(name, m.GetBufferSize(), m.GetPoolSize(), m.GetCacheSize(), m.GetCPUID())
		if err != nil {
			log.Fatalf("Pktmbuf Mempool %s create err: %d", name, err)
		}
		log.Printf("Pktmbuf Mempool %s ready!", name)
	}

	// Add input ports to pipeline
	// pipeline PIPELINE0 port in <portindex> tap <tapname> mempool MEMPOOL0 mtu 1500 bsz 1
	for i, t := range pipelineConfig.InputPorts {
		name := t.GetIfaceName()
		err = dpdki.PipelineAddInputPortTap(pipeName, i, name, t.GetMempool(), t.GetMTU(), t.GetBsz())
		if err != nil {
			log.Fatalf("AddInPortTap %s:%s err: %d", pipeName, name, err)
		}
		log.Printf("AddInPortTap %s:%s ready!", pipeName, name)
	}

	// Add output ports to pipeline
	// pipeline PIPELINE0 port out 0 tap sw0 bsz 1
	for i, t := range pipelineConfig.OutputPorts {
		name := t.GetIfaceName()
		err = dpdki.PipelineAddOutputPortTap(pipeName, i, name, t.GetBsz())
		if err != nil {
			log.Fatalf("AddOutPortTap %s:%s err: %d", pipeName, name, err)
		}
		log.Printf("AddOutPortTap %s:%s ready!", pipeName, name)
	}

	// Build the pipeline program
	err = dpdki.PipelineBuild(pipeName, pipelineConfig.GetSpec())
	if err != nil {
		log.Fatalf("Pipelinebuild %s specfile: %s err: %d", pipeName, pipelineConfig.GetSpec(), err)
	}
	log.Printf("Pipeline %s Build!", pipeName)

	// Commit program to pipeline
	err = dpdki.PipelineCommit(pipeName)
	if err != nil {
		log.Fatalf("Pipelinecommit %s err: %d", pipeName, err)
	}
	log.Printf("Pipeline %s commited!", pipeName)

	// And run pipeline
	// thread 1 pipeline PIPELINE0 enable
	err = dpdki.PipelineEnable(pipeName, pipelineConfig.GetThreadID())
	if err != nil {
		log.Fatalf("PipelineEnable %s err: %d", pipeName, err)
	}
	log.Printf("Pipeline %s enabled!", pipeName)

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
	log.Printf("Table config on pipeline %s commited!", pipeName)
}
