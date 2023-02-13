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
#include <rte_swx_port_ethdev.h>
#include <rte_swx_port_ring.h>
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
	"unsafe"

	"github.com/stolsma/go-p4pack/pkg/dpdkswx"
	"github.com/stolsma/go-p4pack/pkg/dpdkswx/common"
	"github.com/stolsma/go-p4pack/pkg/dpdkswx/swxruntime"
	"github.com/stolsma/go-p4pack/pkg/logging"
)

var log logging.Logger

func init() {
	// keep the logger up to date, also after new log config
	logging.Register("dpdkswx/pipeline", func(logger logging.Logger) {
		log = logger
	})
}

const (
	MaxPortsIn  = 256
	MaxPortsOut = 256
)

type PortParamsType interface {
	PortName() string
	PortType() string
	GetReaderParams() unsafe.Pointer
	GetWriterParams() unsafe.Pointer
	FreeParams()
}

type swxPorts map[int]PortParamsType

// Pipeline represents a DPDK Pipeline record in a Pipeline store
type Pipeline struct {
	Ctl                                 // Pipeline Control Struct inclusion
	name     string                     // Name of the pipeline
	p        *C.struct_rte_swx_pipeline // Struct definition, only for swx internal use!
	build    bool                       // The pipeline is build
	enabled  bool                       // The pipeline is enabled
	threadID uint                       // ID of the Lcore thread this pipeline is running on
	portsIn  swxPorts                   // All added input ports
	portsOut swxPorts                   // All added output ports
	actions  ActionStore                // All the defined actions in this pipeline when build
	tables   TableStore                 // All the defined tables in this pipeline when build
	// TODO mirror slots
	// TODO mirror sessions
	// TODO selectors SelectorStore // All the defined selector tables in this pipeline when build
	learners  LearnerStore  // All the defined learner tables in this pipeline when build
	registers RegisterStore // All the defined registers in this pipeline when build
	meters    MeterStore    // All the defined meters in this pipeline when build
	clean     func()        // The callback function called at clear
}

// Initialize Pipeline. Returns an error if something went wrong.
func (pl *Pipeline) Init(name string, numaNode int, clean func()) error {
	var p *C.struct_rte_swx_pipeline

	// Resource create
	err := dpdkswx.Runtime.ExecOnMain(func(*swxruntime.MainCtx) error {
		status := int(C.rte_swx_pipeline_config(&p, (C.int)(numaNode))) //nolint:gocritic
		if status != 0 {
			C.rte_swx_pipeline_free(p)
			return common.Err(status)
		}
		return nil
	})

	// check if something went wrong when executing on main
	if err != nil {
		return err
	}

	// Node fill in
	pl.name = name
	pl.p = p
	pl.build = false
	pl.enabled = false
	pl.portsIn = make(swxPorts, MaxPortsIn)
	pl.portsOut = make(swxPorts, MaxPortsOut)
	pl.clean = clean

	return nil
}

// Pipeline struct free. If internal pipeline struct pointer is nil, no operation is performed only clean fn is called
// if set in structure.
func (pl *Pipeline) Free() (err error) {
	if pl.p != nil {
		log.Infof("Freeing pipeline: %s", pl.GetName())

		err = dpdkswx.Runtime.ExecOnMain(func(*swxruntime.MainCtx) error {
			if pl.enabled {
				swxruntime.DisablePipeline(pl.GetPipeline())
			}
			pl.Ctl.Free()
			C.rte_swx_pipeline_free(pl.p)
			return nil
		})

		pl.build = false
		pl.enabled = false
		pl.p = nil
	}

	if pl.clean != nil {
		pl.clean()
	}

	return
}

func (pl *Pipeline) GetName() string {
	return pl.name
}

func (pl *Pipeline) GetThreadID() uint {
	return pl.threadID
}

func (pl *Pipeline) GetPipeline() unsafe.Pointer {
	return unsafe.Pointer(pl.p)
}

func (pl *Pipeline) IsBuild() bool {
	return pl.build
}

func (pl *Pipeline) IsEnabled() bool {
	return pl.enabled
}

func (pl *Pipeline) PortInConfig(portID int, params PortParamsType) error {
	if pl.portsIn[portID] != nil {
		return errors.New("port already bound")
	}

	ptype := C.CString(params.PortType())
	defer C.free(unsafe.Pointer(ptype))
	defer params.FreeParams()
	if status := C.rte_swx_pipeline_port_in_config(pl.p, (C.uint)(portID), ptype, params.GetReaderParams()); status != 0 {
		return common.Err(status)
	}

	pl.portsIn[portID] = params

	return nil
}

func (pl *Pipeline) PortOutConfig(portID int, params PortParamsType) error {
	if pl.portsOut[portID] != nil {
		return errors.New("port already bound")
	}

	ptype := C.CString(params.PortType())
	defer C.free(unsafe.Pointer(ptype))
	defer params.FreeParams()
	if status := C.rte_swx_pipeline_port_out_config(pl.p, (C.uint)(portID), ptype, params.GetWriterParams()); status != 0 {
		return common.Err(status)
	}

	pl.portsOut[portID] = params

	return nil
}

func (pl *Pipeline) BuildFromSpec(specfile string) error {
	cspecfile := C.CString(specfile)
	defer C.free(unsafe.Pointer(cspecfile))

	res := C.pipeline_build_from_spec(pl.p, cspecfile)
	if res != 0 {
		return common.Err(res)
	}

	err := pl.Ctl.Init(pl)
	if err != nil {
		return err
	}

	// retrieve actions
	pl.actions = CreateActionStore()
	pl.actions.CreateFromPipeline(pl)

	// retrieve tables
	pl.tables = CreateTableStore()
	pl.tables.CreateFromPipeline(pl)

	// retrieve learner tables
	pl.learners = CreateLearnerStore()
	pl.learners.CreateFromPipeline(pl)

	// retrieve registers
	pl.registers = CreateRegisterStore()
	pl.registers.CreateFromPipeline(pl)

	// retrieve meters
	pl.meters = CreateMeterStore()
	pl.meters.CreateFromPipeline(pl)

	// TODO implement as ENUM state field???
	// pipeline status is build!
	pl.build = true

	return nil
}

// Set pipeline to enabled on given thread
func (pl *Pipeline) SetEnabled(threadID uint) error {
	if pl.enabled {
		return errors.New("pipeline is already enabled")
	}

	err := dpdkswx.Runtime.ExecOnMain(func(*swxruntime.MainCtx) error {
		return swxruntime.EnablePipeline(pl.GetPipeline(), threadID)
	})
	if err != nil {
		return err
	}

	pl.threadID = threadID
	pl.enabled = true

	return nil
}

// Set pipeline to disabled
func (pl *Pipeline) SetDisabled() error {
	if !pl.enabled {
		return errors.New("pipeline is not enabled")
	}

	err := dpdkswx.Runtime.ExecOnMain(func(*swxruntime.MainCtx) error {
		swxruntime.DisablePipeline(pl.GetPipeline())
		return nil
	})
	if err != nil {
		return err
	}

	pl.threadID = 0
	pl.enabled = false

	return nil
}

// Pipeline NUMA node get, return the NUMA node the pipeline is configured to run on.
// Following error codes possible: -EINVAL - Invalid argument.
func (pl *Pipeline) NumaNodeGet() (int, error) {
	var numaNode C.int

	res := C.rte_swx_ctl_pipeline_numa_node_get(pl.p, &numaNode)
	return int(numaNode), common.Err(res)
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
