// SPDX-FileCopyrightText: 2022-present Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package dpdkinfra

import (
	"path"

	"github.com/stolsma/go-p4pack/pkg/dpdkswx/pktmbuf"
)

type PipelineConfig struct {
	Name        string
	NumaNode    int
	BasePath    string
	Spec        string
	ThreadID    uint
	PktMbufs    []*PktMbufConfig
	OutputPorts []*OutPortConfig
	InputPorts  []*InPortConfig
	Start       *StartConfig
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

type PktMbufConfig struct {
	Name       string
	BufferSize uint
	PoolSize   uint32
	CacheSize  uint32
	CPUID      int
}

func (mpc *PktMbufConfig) GetName() string {
	return mpc.Name
}

func (mpc *PktMbufConfig) GetBufferSize() uint {
	if mpc.BufferSize == 0 {
		mpc.BufferSize = pktmbuf.RteMbufDefaultBufSize
	}
	return mpc.BufferSize
}

func (mpc *PktMbufConfig) GetPoolSize() uint32 {
	return mpc.PoolSize
}

func (mpc *PktMbufConfig) GetCacheSize() uint32 {
	return mpc.CacheSize
}

func (mpc *PktMbufConfig) GetCPUID() int {
	return mpc.CPUID
}

type InPortConfig struct {
	IfaceName string
	PktMbuf   string
	MTU       int
	Bsz       int
}

func (pc *InPortConfig) GetIfaceName() string {
	return pc.IfaceName
}

func (pc *InPortConfig) GetPktMbuf() string {
	return pc.PktMbuf
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
func (dpdki *DpdkInfra) PipelineWithConfig(pipelineConfig *PipelineConfig) {
	// Create pipeline
	pipeName := pipelineConfig.GetName()
	err := dpdki.PipelineCreate(pipeName, pipelineConfig.GetNumaNode())
	if err != nil {
		log.Fatalf("%s create err: %d", pipeName, err)
	}
	log.Infof("%s created!", pipeName)

	// Create PktMbuf memory pool
	for _, m := range pipelineConfig.PktMbufs {
		name := m.GetName()
		err := dpdki.PktMbufCreate(name, m.GetBufferSize(), m.GetPoolSize(), m.GetCacheSize(), m.GetCPUID())
		if err != nil {
			log.Fatalf("Pktmbuf Mempool %s create err: %d", name, err)
		}
		log.Infof("Pktmbuf Mempool %s ready!", name)
	}

	// Add input ports to pipeline
	// pipeline PIPELINE0 port in <portindex> tap <tapname> mempool MEMPOOL0 mtu 1500 bsz 1
	for i, t := range pipelineConfig.InputPorts {
		name := t.GetIfaceName()
		err = dpdki.PipelineAddInputPortTap(pipeName, i, name, t.GetPktMbuf(), t.GetMTU(), t.GetBsz())
		if err != nil {
			log.Fatalf("AddInPortTap %s:%s err: %d", pipeName, name, err)
		}
		log.Infof("AddInPortTap %s:%s ready!", pipeName, name)
	}

	// Add output ports to pipeline
	// pipeline PIPELINE0 port out 0 tap sw0 bsz 1
	for i, t := range pipelineConfig.OutputPorts {
		name := t.GetIfaceName()
		err = dpdki.PipelineAddOutputPortTap(pipeName, i, name, t.GetBsz())
		if err != nil {
			log.Fatalf("AddOutPortTap %s:%s err: %d", pipeName, name, err)
		}
		log.Infof("AddOutPortTap %s:%s ready!", pipeName, name)
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
	// thread 1 pipeline PIPELINE0 enable
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
