// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package dpdkswx

/*
#cgo pkg-config: libdpdk

#include <stdlib.h>
#include <string.h>
#include <netinet/in.h>
#include <sys/ioctl.h>
#include <fcntl.h>
#include <unistd.h>
#include <stdint.h>
#include <sys/queue.h>

#include <rte_common.h>
#include <rte_byteorder.h>
#include <rte_cycles.h>
#include <rte_lcore.h>
#include <rte_ring.h>

#include <rte_table_acl.h>
#include <rte_table_array.h>
#include <rte_table_hash.h>
#include <rte_table_lpm.h>
#include <rte_table_lpm_ipv6.h>

#include <rte_mempool.h>
#include <rte_mbuf.h>
#include <rte_ethdev.h>
#include <rte_swx_port_fd.h>
#include <rte_swx_pipeline.h>
#include <rte_swx_ctl.h>
#include <rte_swx_port.h>

int pipeline_build_from_spec(struct rte_swx_pipeline *pipeline, char *specfname) {
	FILE *spec = NULL;
	uint32_t err_line;
	const char *err_msg;
	int status;

	spec = fopen(specfname, "r");
	if (!spec) {
		return 2;
	}

	status = rte_swx_pipeline_build_from_spec(pipeline, spec, &err_line, &err_msg);
	fclose(spec);
	if (status) {
		printf("Err build from spec:%s line: %d\n", err_msg, err_line);
		return status;
	}

	return 0;
}

*/
import "C"
import (
	"errors"
	"fmt"
	"unsafe"
)

// Pipeline represents a DPDK Pipeline record in a Pipeline store
type Pipeline struct {
	PipelineCtl                              // Pipeline Control Struct inclusion
	name          string                     // name of the pipeline
	p             *C.struct_rte_swx_pipeline // Struct definition only swx internal
	timerPeriodms uint32                     //
	build         bool                       // the pipeline is build
	enabled       bool                       // the pipeline is enabled
	threadID      uint32                     // thread pipeline is running on
	cpuID         uint32                     //
	netPortMask   [4]uint64                  //
	actions       ActionStore                // all the defined actions in this pipeline
	tables        TableStore                 // all the defined tables in this pipeline
	clean         func()                     // the callback function called at clear
}

// Initialize Pipeline. Returns an error if something went wrong.
func (pl *Pipeline) Init(name string, numaNode int, clean func()) error {
	var p *C.struct_rte_swx_pipeline

	// Resource create
	status := int(C.rte_swx_pipeline_config(&p, (C.int)(numaNode))) //nolint:gocritic
	if status != 0 {
		C.rte_swx_pipeline_free(p)
		return err(status)
	}

	// Node fill in
	pl.name = name
	pl.p = p
	pl.timerPeriodms = 100
	pl.build = false
	pl.enabled = false
	pl.clean = clean

	return nil
}

func (pl *Pipeline) GetName() string {
	return pl.name
}

func (pl *Pipeline) GetThreadID() uint32 {
	return pl.threadID
}

func (pl *Pipeline) GetPipeline() *C.struct_rte_swx_pipeline {
	return pl.p
}

// Pipeline struct free. If internal pipeline struct pointer is nil, no operation is performed only clean fn is called
// if set in structure.
func (pl *Pipeline) Free() {
	if pl.p != nil {
		pl.PipelineCtl.Free()
		C.rte_swx_pipeline_free(pl.p)
		pl.build = false
		pl.enabled = false
		pl.p = nil
	}

	if pl.clean != nil {
		pl.clean()
	}
}

func (pl *Pipeline) PortInConfig(portID int, portType string, pktmbuf *Pktmbuf, fd C.int, mtu int, bsz int) error {
	var params C.struct_rte_swx_port_fd_reader_params

	if pktmbuf == nil {
		return nil
	}

	params.fd = fd
	params.mempool = pktmbuf.Mempool()
	params.mtu = (C.uint)(mtu)
	params.burst_size = (C.uint)(bsz)
	ptype := C.CString(portType)

	status := C.rte_swx_pipeline_port_in_config(pl.p, (C.uint)(portID), ptype, unsafe.Pointer(&params)) //nolint:gocritic
	C.free(unsafe.Pointer(ptype))

	if status != 0 {
		return err(status)
	}
	return nil
}

// pipeline PIPELINE0 port in 0 tap sw0 mempool MEMPOOL0 mtu 1500 bsz 1
func (pl *Pipeline) AddInputPortTap(portID int, tap *Tap, pktmbuf *Pktmbuf, mtu int, bsz int) error {
	if tap == nil || pktmbuf == nil {
		return nil
	}

	return pl.PortInConfig(portID, "fd", pktmbuf, tap.Fd(), mtu, bsz)
}

func (pl *Pipeline) PortOutConfig(portID int, portType string, fd C.int, bsz int) error {
	var params C.struct_rte_swx_port_fd_writer_params

	params.fd = fd
	params.burst_size = (C.uint)(bsz)
	ptype := C.CString(portType)

	status := int(C.rte_swx_pipeline_port_out_config(pl.p, (C.uint)(portID), ptype, unsafe.Pointer(&params))) //nolint:gocritic
	C.free(unsafe.Pointer(ptype))

	if status != 0 {
		return err(status)
	}
	return nil
}

// pipeline PIPELINE0 port out 0 tap sw0 bsz 1
func (pl *Pipeline) AddOutputPortTap(portID int, tap *Tap, bsz int) error {
	if tap == nil {
		return nil
	}

	return pl.PortOutConfig(portID, "fd", tap.Fd(), bsz)
}

func (pl *Pipeline) BuildFromSpec(specfile string) error {
	cspecfile := C.CString(specfile)
	defer C.free(unsafe.Pointer(cspecfile))

	res := C.pipeline_build_from_spec(pl.p, cspecfile)
	if res != 0 {
		return err(res)
	}

	err := pl.PipelineCtl.Init(pl)
	if err != nil {
		return err
	}

	// retrieve actions
	pl.actions = CreateActionStore()
	pl.actions.CreateFromPipeline(pl)

	// retrieve tables
	pl.tables = CreateTableStore()
	pl.tables.CreateFromPipeline(pl)

	// TODO implement as ENUM state field???
	// pipeline status is build!
	pl.build = true

	return nil
}

func (pl *Pipeline) Enable(threadID uint32) error {
	if pl.enabled {
		return errors.New("pipeline already enabled")
	}

	err := threadPipelineEnable(threadID, pl)
	if err != nil {
		return err
	}

	pl.threadID = threadID
	pl.enabled = true

	return nil
}

func (pl *Pipeline) Disable() error {
	res := threadPipelineDisable(pl)
	pl.enabled = false
	return res
}

// Pipeline NUMA node get
//
// @return int numaNode Pipeline NUMA node.
//
// @return 0 on success or the following error codes otherwise:
// -EINVAL: Invalid argument.
func (pl *Pipeline) NumaNodeGet() (int, error) {
	var numaNode C.int

	res := C.rte_swx_ctl_pipeline_numa_node_get(pl.p, &numaNode)
	return int(numaNode), err(res)
}

type ActionInfo C.struct_rte_swx_ctl_action_info

func (ai *ActionInfo) GetName() string {
	return C.GoString(&ai.name[0])
}

func (ai *ActionInfo) GetNArgs() int {
	return (int)(ai.n_args)
}

func (pl *Pipeline) ActionInfoGet(actionID int) (*ActionInfo, error) {
	var actionInfo = &ActionInfo{}
	result := C.rte_swx_ctl_action_info_get(pl.p, (C.uint)(actionID), (*C.struct_rte_swx_ctl_action_info)(actionInfo))

	if result != 0 {
		return nil, fmt.Errorf("action_info_get error: %d", result)
	}

	return actionInfo, nil
}

type ActionArgInfo C.struct_rte_swx_ctl_action_arg_info

func (aai *ActionArgInfo) GetName() string {
	return C.GoString(&aai.name[0])
}

// Action argument size (in bits)
func (aai *ActionArgInfo) GetNBits() int {
	return (int)(aai.n_bits)
}

// Non-zero (true) when this action argument must be stored in the table in network byte order (NBO), zero when it must
// be stored in host byte order (HBO).
func (aai *ActionArgInfo) IsNetworkByteOrder() bool {
	return aai.is_network_byte_order == 0
}

func (pl *Pipeline) ActionArgInfoGet(actionID int, actionArgID int) (*ActionArgInfo, error) {
	var actionArgInfo = &ActionArgInfo{}
	result := C.rte_swx_ctl_action_arg_info_get(pl.p, (C.uint)(actionID), (C.uint)(actionArgID),
		(*C.struct_rte_swx_ctl_action_arg_info)(actionArgInfo))

	if result != 0 {
		return nil, fmt.Errorf("action_info_get error: %d", result)
	}

	return actionArgInfo, nil
}

type PipelineInfo C.struct_rte_swx_ctl_pipeline_info

func (pi *PipelineInfo) GetNPortsIn() int {
	return int(pi.n_ports_in)
}

func (pi *PipelineInfo) GetNPortsOut() int {
	return int(pi.n_ports_out)
}

func (pi *PipelineInfo) GetNMirroringSlots() int {
	return int(pi.n_mirroring_slots)
}

func (pi *PipelineInfo) GetNMirroringSessions() int {
	return int(pi.n_mirroring_sessions)
}

func (pi *PipelineInfo) GetNActions() int {
	return int(pi.n_actions)
}

func (pi *PipelineInfo) GetNTables() int {
	return int(pi.n_tables)
}

func (pi *PipelineInfo) GetNSelectors() int {
	return int(pi.n_selectors)
}

func (pi *PipelineInfo) GetNLearners() int {
	return int(pi.n_learners)
}

func (pi *PipelineInfo) GetNRegarrays() int {
	return int(pi.n_regarrays)
}

func (pi *PipelineInfo) GetNMetarrays() int {
	return int(pi.n_metarrays)
}

func (pl *Pipeline) PipelineInfoGet() (*PipelineInfo, error) {
	var pipeInfo PipelineInfo

	res := C.rte_swx_ctl_pipeline_info_get(pl.p, (*C.struct_rte_swx_ctl_pipeline_info)(&pipeInfo))
	if res < 0 {
		return nil, errors.New("pipelineInfoGet failed")
	}

	return &pipeInfo, nil
}

// information about the structure of a table
type TableInfo C.struct_rte_swx_ctl_table_info

// return the name of the table
func (ti *TableInfo) GetName() string {
	return C.GoString(&ti.name[0])
}

func (ti *TableInfo) GetArgs() string {
	return C.GoString(&ti.args[0])
}

func (ti *TableInfo) GetNMatchFields() int {
	return int(ti.n_match_fields)
}

func (ti *TableInfo) GetNActions() int {
	return int(ti.n_actions)
}

func (ti *TableInfo) GetDefaultActionIsConst() bool {
	return ti.default_action_is_const > 0
}

func (ti *TableInfo) GetSize() int {
	return int(ti.size)
}

func (pl *Pipeline) TableInfoGet(tableID int) (*TableInfo, error) {
	var tableInfo TableInfo

	status := C.rte_swx_ctl_table_info_get(pl.p, (C.uint)(tableID), (*C.struct_rte_swx_ctl_table_info)(&tableInfo))
	if status != 0 {
		return nil, fmt.Errorf("Table (ID: %d) info get error", tableID)
	}
	return &tableInfo, nil
}

// information about table match fields
type TableMatchFieldInfo C.struct_rte_swx_ctl_table_match_field_info

func (tmfi *TableMatchFieldInfo) GetMatchType() int {
	return int(tmfi.match_type)
}

func (tmfi *TableMatchFieldInfo) GetIsHeader() bool {
	return tmfi.is_header > 0
}

func (tmfi *TableMatchFieldInfo) GetNBits() int {
	return int(tmfi.n_bits)
}

func (tmfi *TableMatchFieldInfo) GetOffset() int {
	return int(tmfi.offset)
}

func (pl *Pipeline) TableMatchFieldInfoGet(tableID int, matchFieldID int) (*TableMatchFieldInfo, error) {
	var tableMatchFieldInfo TableMatchFieldInfo

	status := C.rte_swx_ctl_table_match_field_info_get(
		pl.p, (C.uint)(tableID), (C.uint)(matchFieldID), (*C.struct_rte_swx_ctl_table_match_field_info)(&tableMatchFieldInfo))
	if status != 0 {
		return nil, fmt.Errorf("Table (ID: %d) match field (ID: %d) info get error", tableID, matchFieldID)
	}
	return &tableMatchFieldInfo, nil
}

// information about table actions
type TableActionInfo C.struct_rte_swx_ctl_table_action_info

func (tai *TableActionInfo) GetActionID() int {
	return int(tai.action_id)
}

func (tai *TableActionInfo) GetActionIsForTableEntries() bool {
	return tai.action_is_for_table_entries > 0
}

func (tai *TableActionInfo) GetActionIsForDefaultEntry() bool {
	return tai.action_is_for_default_entry > 0
}

func (pl *Pipeline) TableActionInfoGet(tableID int, actionID int) (*TableActionInfo, error) {
	var tableActionInfo TableActionInfo

	status := C.rte_swx_ctl_table_action_info_get(
		pl.p, (C.uint)(tableID), (C.uint)(actionID), (*C.struct_rte_swx_ctl_table_action_info)(&tableActionInfo))
	if status != 0 {
		return nil, fmt.Errorf("Table (ID: %d) action (ID: %d) info get error", tableID, actionID)
	}
	return &tableActionInfo, nil
}

//
// Statistic functions
//

type PortInStats C.struct_rte_swx_port_in_stats

func (pl *Pipeline) PortInStatsRead(port int) (*PortInStats, error) {
	var portInStats PortInStats

	C.rte_swx_ctl_pipeline_port_in_stats_read(pl.p, (C.uint)(port), (*C.struct_rte_swx_port_in_stats)(&portInStats))
	return &portInStats, nil
}

type PortOutStats C.struct_rte_swx_port_out_stats

func (pl *Pipeline) PortOutStatsRead(port int) (*PortOutStats, error) {
	var portOutStats PortOutStats

	C.rte_swx_ctl_pipeline_port_out_stats_read(pl.p, (C.uint)(port), (*C.struct_rte_swx_port_out_stats)(&portOutStats))
	return &portOutStats, nil
}

type TableStats struct {
	nPktsHit    uint64
	nPktsMiss   uint64
	nPktsAction []actionfield
}

type actionfield struct {
	name string
	pkts uint64
}

func (pl *Pipeline) TableStatsRead(tableName string) (*TableStats, error) {
	actionSize := len(pl.actions)
	ret := C.malloc(C.size_t(actionSize) * C.size_t(unsafe.Sizeof(C.uint64_t(0))))
	defer C.free(ret)
	cTableName := C.CString(tableName)
	defer C.free(unsafe.Pointer(cTableName))

	var cTableStats = C.struct_rte_swx_table_stats{
		n_pkts_hit:    0,
		n_pkts_miss:   0,
		n_pkts_action: (*C.uint64_t)(ret),
	}

	status := C.rte_swx_ctl_pipeline_table_stats_read(pl.p, cTableName, &cTableStats) //nolint:gocritic
	if status != 0 {
		return nil, fmt.Errorf("Table (Name: %s) stats read error", tableName)
	}

	table := pl.tables.Find(tableName)
	var tableStats = TableStats{
		nPktsHit:    uint64(cTableStats.n_pkts_hit),
		nPktsMiss:   uint64(cTableStats.n_pkts_miss),
		nPktsAction: make([]actionfield, table.nActions),
	}
	actionStat := unsafe.Slice((*uint64)(ret), actionSize)
	table.actions.ForEach(func(action *TableAction) {
		tableStats.nPktsAction[action.GetIndex()] = actionfield{
			name: action.GetName(),
			pkts: actionStat[action.action.GetIndex()],
		}
	})

	return &tableStats, nil
}

func isPowerOfTwo(x int) bool {
	return (x != 0) && ((x & (x - 1)) == 0)
}

// Validate the number of ports added to the pipeline in input and output directions
func (pl *Pipeline) PortIsValid() bool {
	pipeInfo, err := pl.PipelineInfoGet()
	if err != nil {
		return false
	}

	if pipeInfo.n_ports_in == 0 || !(isPowerOfTwo((int)(pipeInfo.n_ports_in))) {
		return false
	}

	if pipeInfo.n_ports_out == 0 {
		return false
	}
	return true
}

const infoTemplate = `	ports in           : %d
	ports out          : %d
	mirroring slots    : %d
	mirroring sessions : %d
	actions            : %d
	tables             : %d
	selectors          : %d
	learners           : %d
	register arrays    : %d
	meta arrays        : %d
`

func (pl *Pipeline) Info() string {
	pi, err := pl.PipelineInfoGet()
	if err != nil {
		return ""
	}

	result := fmt.Sprintf(infoTemplate, pi.n_ports_in, pi.n_ports_out, pi.n_mirroring_slots, pi.n_mirroring_sessions,
		pi.n_actions, pi.n_tables, pi.n_selectors, pi.n_learners, pi.n_regarrays, pi.n_metarrays)

	return result
}

func (pl *Pipeline) Stats() string {
	var result string

	pipeInfo, err := pl.PipelineInfoGet()
	if err != nil {
		return ""
	}

	result += "Input ports:\n"
	for i := 0; i < (int)(pipeInfo.n_ports_in); i++ {
		portInStats, err := pl.PortInStatsRead(i)
		if err != nil {
			return ""
		}
		result += fmt.Sprintf("\tPort %d\t Packets: %d\tBytes: %d\tEmpty: %d\n",
			i, portInStats.n_pkts, portInStats.n_bytes, portInStats.n_empty)
	}

	result += "\nOutput ports:\n"
	for i := 0; i < (int)(pipeInfo.n_ports_out); i++ {
		portOutStats, err := pl.PortOutStatsRead(i)
		if err != nil {
			return ""
		}

		if i != (int)(pipeInfo.n_ports_out)-1 {
			result += fmt.Sprintf("\tPort %d\t Packets: %d\tBytes: %d\tClone: %d\tClone Error: %d\n",
				i, portOutStats.n_pkts, portOutStats.n_bytes, portOutStats.n_pkts_clone, portOutStats.n_pkts_clone_err)
		} else {
			result += fmt.Sprintf("\tDROP\t Packets: %d\tBytes: %d\n", portOutStats.n_pkts, portOutStats.n_bytes)
		}
	}

	result += "\nTables:\n"
	for i := 0; i < (int)(pipeInfo.n_tables); i++ {
		tableInfo, err := pl.TableInfoGet(i)
		if err != nil {
			return ""
		}

		tableStats, err := pl.TableStatsRead(tableInfo.GetName())
		if err != nil {
			return ""
		}

		result += fmt.Sprintf("\tTable %s:\n", tableInfo.GetName())
		result += fmt.Sprintf("\t\tHit (packets) : %d\n", tableStats.nPktsHit)
		result += fmt.Sprintf("\t\tMiss (packets): %d\n", tableStats.nPktsMiss)

		for i := 0; i < len(tableStats.nPktsAction); i++ {
			result += fmt.Sprintf("\t\t%s action (packets): %d\n", tableStats.nPktsAction[i].name, tableStats.nPktsAction[i].pkts)
		}
	}

	return result

	/*

		for (i = 0; i < info.n_tables; i++) {
			struct rte_swx_ctl_table_info table_info;
			uint64_t n_pkts_action[info.n_actions];
			struct rte_swx_table_stats stats = {
				.n_pkts_hit = 0,
				.n_pkts_miss = 0,
				.n_pkts_action = n_pkts_action,
			};
			uint32_t j;

			status = rte_swx_ctl_table_info_get(p->p, i, &table_info);
			if (status) {
				snprintf(out, out_size, "Table info get error.");
				return;
			}

			status = rte_swx_ctl_pipeline_table_stats_read(p->p, table_info.name, &stats);
			if (status) {
				snprintf(out, out_size, "Table stats read error.");
				return;
			}

			snprintf(out, out_size, "\tTable %s:\n"
				"\t\tHit (packets): %" PRIu64 "\n"
				"\t\tMiss (packets): %" PRIu64 "\n",
				table_info.name,
				stats.n_pkts_hit,
				stats.n_pkts_miss);
			out_size -= strlen(out);
			out += strlen(out);

			for (j = 0; j < info.n_actions; j++) {
				struct rte_swx_ctl_action_info action_info;

				status = rte_swx_ctl_action_info_get(p->p, j, &action_info);
				if (status) {
					snprintf(out, out_size, "Action info get error.");
					return;
				}

				snprintf(out, out_size, "\t\tAction %s (packets): %" PRIu64 "\n",
					action_info.name,
					stats.n_pkts_action[j]);
				out_size -= strlen(out);
				out += strlen(out);
			}
		}

		snprintf(out, out_size, "\nLearner tables:\n");
		out_size -= strlen(out);
		out += strlen(out);

		for (i = 0; i < info.n_learners; i++) {
			struct rte_swx_ctl_learner_info learner_info;
			uint64_t n_pkts_action[info.n_actions];
			struct rte_swx_learner_stats stats = {
				.n_pkts_hit = 0,
				.n_pkts_miss = 0,
				.n_pkts_action = n_pkts_action,
			};
			uint32_t j;

			status = rte_swx_ctl_learner_info_get(p->p, i, &learner_info);
			if (status) {
				snprintf(out, out_size, "Learner table info get error.");
				return;
			}

			status = rte_swx_ctl_pipeline_learner_stats_read(p->p, learner_info.name, &stats);
			if (status) {
				snprintf(out, out_size, "Learner table stats read error.");
				return;
			}

			snprintf(out, out_size, "\tLearner table %s:\n"
				"\t\tHit (packets): %" PRIu64 "\n"
				"\t\tMiss (packets): %" PRIu64 "\n"
				"\t\tLearn OK (packets): %" PRIu64 "\n"
				"\t\tLearn error (packets): %" PRIu64 "\n"
				"\t\tForget (packets): %" PRIu64 "\n",
				learner_info.name,
				stats.n_pkts_hit,
				stats.n_pkts_miss,
				stats.n_pkts_learn_ok,
				stats.n_pkts_learn_err,
				stats.n_pkts_forget);
			out_size -= strlen(out);
			out += strlen(out);

			for (j = 0; j < info.n_actions; j++) {
				struct rte_swx_ctl_action_info action_info;

				status = rte_swx_ctl_action_info_get(p->p, j, &action_info);
				if (status) {
					snprintf(out, out_size, "Action info get error.");
					return;
				}

				snprintf(out, out_size, "\t\tAction %s (packets): %" PRIu64 "\n",
					action_info.name,
					stats.n_pkts_action[j]);
				out_size -= strlen(out);
				out += strlen(out);
			}
		}

	*/
}

/*
uint64_t get_action_id(struct pipeline *pipe, const char *action_name) {
	uint64_t i;
	int ret;
	struct rte_swx_ctl_action_info action;
	struct rte_swx_ctl_pipeline_info pipe_info = {0};

	if (action_name == NULL || pipe == NULL || pipe->p == NULL) {
		printf("%s failed at %d\n",__func__, __LINE__);
		goto action_error;
	}
	ret = rte_swx_ctl_pipeline_info_get(pipe->p, &pipe_info);
	if (ret < 0) {
		printf("%s failed at %d for pipeinfo \n",__func__, __LINE__);
		goto action_error;
	}
	for (i = 0; i < pipe_info.n_actions; i++) {
		memset(&action, 0, sizeof(action));
		ret = rte_swx_ctl_action_info_get (pipe->p, i, &action);
		if (ret < 0) {
			printf("%s failed at %d for actioninfo\n",
				__func__, __LINE__);
			break;
		}
		if (!strncmp(action_name, action.name, RTE_SWX_CTL_NAME_SIZE))
			return i;
	}
action_error:
	printf("%s failed at %d end\n",__func__, __LINE__);
	return UINT64_MAX;
}

uint32_t get_table_id(struct pipeline *pipe, const char *table_name) {
	uint32_t i;
	int ret;
	struct rte_swx_ctl_table_info table;
	struct rte_swx_ctl_pipeline_info pipe_info = {0};

	if (table_name == NULL || pipe == NULL || pipe->p == NULL) {
		printf("%s failed at %d\n",__func__, __LINE__);
		goto table_error;
	}

	ret = rte_swx_ctl_pipeline_info_get(pipe->p, &pipe_info);
	if (ret < 0) {
		printf("%s failed at %d for pipeinfo\n",__func__, __LINE__);
		goto table_error;
	}
	for (i = 0; i < pipe_info.n_tables; i++) {
		memset(&table, 0, sizeof(table));
		ret = rte_swx_ctl_table_info_get (pipe->p, i, &table);
		if (ret < 0) {
			printf("%s failed at %d for tableinfo\n",
				__func__, __LINE__);
			break;
		}
		if (!strncmp(table_name, table.name, RTE_SWX_CTL_NAME_SIZE))
			return i;
	}
table_error:
	printf("%s failed at %d end\n",__func__, __LINE__);
	return UINT32_MAX;
}

*/
