// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package pipeline

/*
#include <stdlib.h>
#include <string.h>
#include <netinet/in.h>
#include <sys/ioctl.h>
#include <fcntl.h>
#include <unistd.h>
#include <stdint.h>

#include <rte_mempool.h>
#include <rte_mbuf.h>
#include <rte_ethdev.h>
#include <rte_swx_port_fd.h>
#include <rte_swx_port_ethdev.h>
#include <rte_swx_port_ring.h>
#include <rte_swx_pipeline.h>
#include <rte_swx_ctl.h>
#include <rte_swx_port.h>

*/
import "C"

import (
	"fmt"
	"unsafe"
)

type PortInStats C.struct_rte_swx_port_in_stats

// Single line port in statistics
func (pis *PortInStats) String() string {
	return fmt.Sprintf("Packets: %-20d Bytes: %-20d Empty: %-20d", pis.n_pkts, pis.n_bytes, pis.n_empty)
}

func (pl *Pipeline) PortInStatsRead(port int) (*PortInStats, error) {
	var portInStats PortInStats

	C.rte_swx_ctl_pipeline_port_in_stats_read(pl.p, (C.uint)(port), (*C.struct_rte_swx_port_in_stats)(&portInStats))
	return &portInStats, nil
}

type PortOutStats C.struct_rte_swx_port_out_stats

// Single line port out statistics
func (pos *PortOutStats) String() string {
	result := fmt.Sprintf("Packets: %-20d Bytes: %-20d", pos.n_pkts, pos.n_bytes)

	if pos.n_pkts_clone > 0 {
		result += fmt.Sprintf(" Clone: %-20d", pos.n_pkts_clone)
	}

	if pos.n_pkts_clone_err > 0 {
		result += fmt.Sprintf(" Clone Error: %-20d", pos.n_pkts_clone_err)
	}

	return result
}

func (pl *Pipeline) PortOutStatsRead(port int) (*PortOutStats, error) {
	var portOutStats PortOutStats

	C.rte_swx_ctl_pipeline_port_out_stats_read(pl.p, (C.uint)(port), (*C.struct_rte_swx_port_out_stats)(&portOutStats))
	return &portOutStats, nil
}

type ActionFieldStat struct {
	name string
	pkts uint64
}

func (af *ActionFieldStat) GetName() string {
	return af.name
}

// Single line action field statistics
func (af *ActionFieldStat) String() string {
	return fmt.Sprintf("%s (packets): %-20d", af.name, af.pkts)
}

type TableStats struct {
	name        string
	nPktsHit    uint64
	nPktsMiss   uint64
	nPktsAction []ActionFieldStat
}

func (ts *TableStats) GetName() string {
	return ts.name
}

// Multi line table statistics
func (ts *TableStats) String() string {
	result := fmt.Sprintf("Hit (packets) : %-20d\n", ts.nPktsHit)
	result += fmt.Sprintf("Miss (packets): %-20d\n", ts.nPktsMiss)

	for i := 0; i < len(ts.nPktsAction); i++ {
		result += ts.nPktsAction[i].String() + "\n"
	}

	return result
}

func (pl *Pipeline) TableStatsRead(tableName string) (*TableStats, error) {
	actionSize := len(pl.actions)
	cPktsAction := C.malloc(C.size_t(actionSize) * C.size_t(unsafe.Sizeof(C.uint64_t(0))))
	defer C.free(cPktsAction)
	cTableName := C.CString(tableName)
	defer C.free(unsafe.Pointer(cTableName))

	var cTableStats = C.struct_rte_swx_table_stats{
		n_pkts_hit:    0,
		n_pkts_miss:   0,
		n_pkts_action: (*C.uint64_t)(cPktsAction),
	}

	status := C.rte_swx_ctl_pipeline_table_stats_read(pl.p, cTableName, &cTableStats) //nolint:gocritic
	if status != 0 {
		return nil, fmt.Errorf("Table (Name: %s) stats read error", tableName)
	}

	table := pl.tables.FindName(tableName)
	var tableStats = &TableStats{
		name:        tableName,
		nPktsHit:    uint64(cTableStats.n_pkts_hit),
		nPktsMiss:   uint64(cTableStats.n_pkts_miss),
		nPktsAction: make([]ActionFieldStat, table.nActions),
	}
	actionStat := unsafe.Slice((*uint64)(cPktsAction), actionSize) // cast back from C structure
	table.actions.ForEach(func(key string, action *TableAction) error {
		tableStats.nPktsAction[action.GetIndex()] = ActionFieldStat{
			name: action.GetActionName(),
			pkts: actionStat[action.action.GetIndex()],
		}
		return nil
	})

	return tableStats, nil
}

type LearnerStats struct {
	name          string
	nPktsHit      uint64
	nPktsMiss     uint64
	nPktsLearnOk  uint64
	nPktsLearnErr uint64
	nPktsRearm    uint64
	nPktsForget   uint64
	nPktsAction   []ActionFieldStat
}

func (ls *LearnerStats) GetName() string {
	return ls.name
}

// Multi line learner table statistics
func (ls *LearnerStats) String() string {
	result := fmt.Sprintf("Hit (packets)         : %-20d\n", ls.nPktsHit)
	result += fmt.Sprintf("Miss (packets)        : %-20d\n", ls.nPktsMiss)
	result += fmt.Sprintf("Learn OK (packets)    : %-20d\n", ls.nPktsLearnOk)
	result += fmt.Sprintf("Learn Error (packets) : %-20d\n", ls.nPktsLearnErr)
	result += fmt.Sprintf("Rearm (packets)       : %-20d\n", ls.nPktsRearm)
	result += fmt.Sprintf("Forget (packets)      : %-20d\n", ls.nPktsForget)

	for i := 0; i < len(ls.nPktsAction); i++ {
		result += ls.nPktsAction[i].String() + "\n"
	}

	return result
}

func (pl *Pipeline) LearnerStatsRead(tableName string) (*LearnerStats, error) {
	actionSize := len(pl.actions)
	cPktsAction := C.malloc(C.size_t(actionSize) * C.size_t(unsafe.Sizeof(C.uint64_t(0))))
	defer C.free(cPktsAction)
	cTableName := C.CString(tableName)
	defer C.free(unsafe.Pointer(cTableName))

	var cLearnerStats = C.struct_rte_swx_learner_stats{
		n_pkts_hit:       0,
		n_pkts_miss:      0,
		n_pkts_learn_ok:  0,
		n_pkts_learn_err: 0,
		n_pkts_rearm:     0,
		n_pkts_forget:    0,
		n_pkts_action:    (*C.uint64_t)(cPktsAction),
	}

	status := C.rte_swx_ctl_pipeline_learner_stats_read(pl.p, cTableName, &cLearnerStats) //nolint:gocritic
	if status != 0 {
		return nil, fmt.Errorf("Table (Name: %s) stats read error", tableName)
	}

	table := pl.tables.FindName(tableName)
	var LearnerStats = LearnerStats{
		name:          tableName,
		nPktsHit:      uint64(cLearnerStats.n_pkts_hit),
		nPktsMiss:     uint64(cLearnerStats.n_pkts_miss),
		nPktsLearnOk:  uint64(cLearnerStats.n_pkts_learn_ok),
		nPktsLearnErr: uint64(cLearnerStats.n_pkts_learn_err),
		nPktsRearm:    uint64(cLearnerStats.n_pkts_rearm),
		nPktsForget:   uint64(cLearnerStats.n_pkts_forget),
		nPktsAction:   make([]ActionFieldStat, table.nActions),
	}
	actionStat := unsafe.Slice((*uint64)(cPktsAction), actionSize) // cast back from C structure
	table.actions.ForEach(func(key string, action *TableAction) error {
		LearnerStats.nPktsAction[action.GetIndex()] = ActionFieldStat{
			name: action.GetActionName(),
			pkts: actionStat[action.action.GetIndex()],
		}
		return nil
	})

	return &LearnerStats, nil
}

func isPowerOfTwo(x int) bool {
	return (x != 0) && ((x & (x - 1)) == 0)
}

type Stats struct {
	PortInStats  []*PortInStats
	PortOutStats []*PortOutStats
	TableStats   []*TableStats
	LearnerStats []*LearnerStats
}

func (pls *Stats) String() string {
	result := "Input ports:\n"
	for i, pis := range pls.PortInStats {
		result += fmt.Sprintf("Port %-3d %s\n", i, pis.String())
	}

	result += "\nOutput ports:\n"
	for i, pos := range pls.PortOutStats {
		if i != len(pls.PortOutStats)-1 {
			result += fmt.Sprintf("Port %-3d %s\n", i, pos.String())
		} else {
			result += fmt.Sprintf("DROP     %s\n", pos.String())
		}
	}

	result += "\nTables:\n"
	for _, ts := range pls.TableStats {
		result += fmt.Sprintf("Table %s:\n", ts.GetName())
		result += ts.String()
	}

	result += "\nLearner Tables:\n"
	for _, ls := range pls.LearnerStats {
		result += fmt.Sprintf("Table %s:\n", ls.GetName())
		result += ls.String()
	}

	return result
}

func (pl *Pipeline) StatsRead() (pls *Stats, err error) {
	pls = &Stats{}

	pipeInfo, err := pl.PipelineInfoGet()
	if err != nil {
		return nil, err
	}

	// get port in stats
	pls.PortInStats = make([]*PortInStats, pipeInfo.n_ports_in)
	for i := 0; i < (int)(pipeInfo.n_ports_in); i++ {
		portInStats, err := pl.PortInStatsRead(i)
		if err != nil {
			return nil, err
		}
		pls.PortInStats[i] = portInStats
	}

	// get port out stats
	pls.PortOutStats = make([]*PortOutStats, pipeInfo.n_ports_out)
	for i := 0; i < (int)(pipeInfo.n_ports_out); i++ {
		portOutStats, err := pl.PortOutStatsRead(i)
		if err != nil {
			return nil, err
		}
		pls.PortOutStats[i] = portOutStats
	}

	// get table stats
	pls.TableStats = make([]*TableStats, pipeInfo.n_tables)
	for i := 0; i < (int)(pipeInfo.n_tables); i++ {
		tableInfo, err := pl.TableInfoGet(i)
		if err != nil {
			return nil, err
		}

		tableStats, err := pl.TableStatsRead(tableInfo.GetName())
		if err != nil {
			return nil, err
		}
		pls.TableStats[i] = tableStats
	}

	// get learner table stats
	pls.LearnerStats = make([]*LearnerStats, pipeInfo.n_learners)
	for i := 0; i < (int)(pipeInfo.n_learners); i++ {
		li, err := pl.LearnerInfoGet(i)
		if err != nil {
			return nil, err
		}

		learnerStats, err := pl.LearnerStatsRead(li.GetName())
		if err != nil {
			return nil, err
		}

		pls.LearnerStats[i] = learnerStats
	}

	return pls, nil
}
