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

int pipeline_build(struct rte_swx_pipeline *pipeline, char *specfname) {
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
	name          string
	p             *C.struct_rte_swx_pipeline     // Struct definition only swx internal
	ctl           *C.struct_rte_swx_ctl_pipeline // Struct definition only swx internal
	timerPeriodms uint32                         //
	build         bool                           // the pipeline is build
	enabled       bool                           // the pipeline is enabled
	thread_id     uint32                         // thread pipeline is running on
	cpu_id        uint32                         //
	net_port_mask [4]uint64                      //
	clean         func()                         // the callback function called at clear
}

// Create Pipeline. Returns a pointer to a Pipeline structure or nil with error.
func (pl *Pipeline) Init(name string, numaNode int, clean func()) error {
	var p *C.struct_rte_swx_pipeline

	// Resource create
	status := C.rte_swx_pipeline_config(&p, (C.int)(numaNode))
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

func (pl *Pipeline) Name() string {
	return pl.name
}

func (pl *Pipeline) ThreadId() uint32 {
	return pl.thread_id
}

func (pl *Pipeline) Pipeline() *C.struct_rte_swx_pipeline {
	return pl.p
}

func (pl *Pipeline) Free() {
	if pl.p != nil {
		//		C.rte_mempool_free((*C.struct_rte_mempool)(pm.m))
		pl.p = nil
	}

	if pl.clean != nil {
		pl.clean()
	}
}

// pipeline PIPELINE0 port in 0 tap sw0 mempool MEMPOOL0 mtu 1500 bsz 1
func (pl *Pipeline) AddInputPortTap(portId int, tap *Tap, pktmbuf *Pktmbuf, mtu int, bsz int) error {
	var params C.struct_rte_swx_port_fd_reader_params

	if tap == nil {
		return nil
	}

	if pktmbuf == nil {
		return nil
	}

	params.fd = tap.Fd()
	params.mempool = pktmbuf.Mempool()
	params.mtu = (C.uint)(mtu)
	params.burst_size = (C.uint)(bsz)
	ptype := C.CString("fd")

	status := C.rte_swx_pipeline_port_in_config(pl.p, (C.uint)(portId), ptype, unsafe.Pointer(&params))
	C.free(unsafe.Pointer(ptype))

	if status != 0 {
		return err(status)
	}
	return nil
}

// pipeline PIPELINE0 port out 0 tap sw0 bsz 1
func (pl *Pipeline) AddOutputPortTap(portId int, tap *Tap, bsz int) error {
	var params C.struct_rte_swx_port_fd_writer_params

	if tap == nil {
		return nil
	}

	params.fd = tap.Fd()
	params.burst_size = (C.uint)(bsz)
	ptype := C.CString("fd")

	status := C.rte_swx_pipeline_port_out_config(pl.p, (C.uint)(portId), ptype, unsafe.Pointer(&params))
	C.free(unsafe.Pointer(ptype))

	if status != 0 {
		return err(status)
	}
	return nil
}

func (pl *Pipeline) Build(specfile string) error {
	cspecfile := C.CString(specfile)
	defer C.free(unsafe.Pointer(cspecfile))

	res := C.pipeline_build(pl.p, cspecfile)
	if res != 0 {
		return err(res)
	}

	pctl := C.rte_swx_ctl_pipeline_create(pl.p)
	if pctl == nil {
		//rte_swx_pipeline_free(pipeline)
		return errors.New("rte_swx_ctl_pipeline_create error")
	}
	pl.ctl = pctl
	pl.build = true

	return nil
}

func (pl *Pipeline) Commit() error {
	var abortOnFail C.int = 1
	res := C.rte_swx_ctl_pipeline_commit(pl.ctl, abortOnFail)
	return err(res)
}

func (pl *Pipeline) Enable(threadId uint32) error {
	if pl.enabled {
		return errors.New("pipeline already enabled")
	}

	err := threadPipelineEnable(threadId, pl)
	if err != nil {
		return err
	}

	pl.thread_id = threadId
	pl.enabled = true

	return nil
}

func (pl *Pipeline) Disable() error {
	res := threadPipelineDisable(pl)
	pl.enabled = false
	return res
}

//
// Statistic functions
//

type PipelineInfo C.struct_rte_swx_ctl_pipeline_info

func (pl *Pipeline) pipelineInfo() (*PipelineInfo, error) {
	var pipeInfo PipelineInfo

	res := C.rte_swx_ctl_pipeline_info_get(pl.p, (*C.struct_rte_swx_ctl_pipeline_info)(&pipeInfo))
	if res < 0 {
		return nil, errors.New("pipelineInfo failed")
	}

	return &pipeInfo, nil
}

type PortInStats C.struct_rte_swx_port_in_stats

func (pl *Pipeline) PortInStats(port int) (*PortInStats, error) {
	var portInStats PortInStats

	C.rte_swx_ctl_pipeline_port_in_stats_read(pl.p, (C.uint)(port), (*C.struct_rte_swx_port_in_stats)(&portInStats))
	return &portInStats, nil
}

type PortOutStats C.struct_rte_swx_port_out_stats

func (pl *Pipeline) PortOutStats(port int) (*PortOutStats, error) {
	var portOutStats PortOutStats

	C.rte_swx_ctl_pipeline_port_out_stats_read(pl.p, (C.uint)(port), (*C.struct_rte_swx_port_out_stats)(&portOutStats))
	return &portOutStats, nil
}

type TableStats C.struct_rte_swx_table_stats

func (pl *Pipeline) TableStats(tableId uint) (*TableStats, error) {

	//	_, err := GetTable(pl, tableId)
	//	if err != nil {
	//		return nil, err
	//	}

	return nil, nil
	// TODO implement
	/*
		var  table_info C.struct_rte_swx_ctl_table_info

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
	*/
}

func isPowerOfTwo(x int) bool {
	return (x != 0) && ((x & (x - 1)) == 0)
}

// Validate the number of ports added to the pipeline in input and output directions
func (pl *Pipeline) PortIsValid() bool {
	pipeInfo, err := pl.pipelineInfo()
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

func (pl *Pipeline) Stats() string {
	var result string = ""

	pipeInfo, err := pl.pipelineInfo()
	if err != nil {
		return ""
	}

	result += "Input ports:\n"
	for i := 0; i < (int)(pipeInfo.n_ports_in); i++ {
		portInStats, err := pl.PortInStats(i)
		if err != nil {
			return ""
		}
		result += fmt.Sprintf("\tPort %d\t Packets: %d\tBytes: %d\tEmpty: %d\n",
			i, portInStats.n_pkts, portInStats.n_bytes, portInStats.n_empty)
	}

	result += "\nOutput ports:\n"
	for i := 0; i < (int)(pipeInfo.n_ports_out); i++ {
		portOutStats, err := pl.PortInStats(i)
		if err != nil {
			return ""
		}

		if i != (int)(pipeInfo.n_ports_out)-1 {
			result += fmt.Sprintf("\tPort %d\t Packets: %d\tBytes: %d\tEmpty: %d\n",
				i, portOutStats.n_pkts, portOutStats.n_bytes, portOutStats.n_empty)
		} else {
			result += fmt.Sprintf("\tDROP\t Packets: %d\tBytes: %d\tEmpty: %d\n",
				portOutStats.n_pkts, portOutStats.n_bytes, portOutStats.n_empty)
		}
	}

	result += "\nTables:\n"
	for i := 0; i < (int)(pipeInfo.n_tables); i++ {
		//		GetTable(pl, (uint)(i))
		/*		portInStats, err := pl.PortInStats(i)
				if err != nil {
					return ""
				}
				result += fmt.Sprintf("\tPort %d\t Packets: %d\tBytes: %d\tEmpty: %d\n",
					i, portInStats.n_pkts, portInStats.n_bytes, portInStats.n_empty)
		*/
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

void table_entry_free(struct rte_swx_table_entry *entry) {
	if (!entry)
		return;

	free(entry->key);
	free(entry->key_mask);
	free(entry->action_data);
	free(entry);
}

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
