// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package pipeline

/*
#include <rte_swx_pipeline.h>
#include <rte_swx_ctl.h>
*/
import "C"
import (
	"errors"
	"fmt"
)

type Info C.struct_rte_swx_ctl_pipeline_info

func (pi *Info) GetNPortsIn() int {
	return int(pi.n_ports_in)
}

func (pi *Info) GetNPortsOut() int {
	return int(pi.n_ports_out)
}

func (pi *Info) GetNMirroringSlots() int {
	return int(pi.n_mirroring_slots)
}

func (pi *Info) GetNMirroringSessions() int {
	return int(pi.n_mirroring_sessions)
}

func (pi *Info) GetNActions() int {
	return int(pi.n_actions)
}

func (pi *Info) GetNTables() int {
	return int(pi.n_tables)
}

func (pi *Info) GetNSelectors() int {
	return int(pi.n_selectors)
}

func (pi *Info) GetNLearners() int {
	return int(pi.n_learners)
}

func (pi *Info) GetNRegarrays() int {
	return int(pi.n_regarrays)
}

func (pi *Info) GetNMetarrays() int {
	return int(pi.n_metarrays)
}

func (pl *Pipeline) PipelineInfoGet() (*Info, error) {
	var pipeInfo Info

	res := C.rte_swx_ctl_pipeline_info_get(pl.p, (*C.struct_rte_swx_ctl_pipeline_info)(&pipeInfo))
	if res < 0 {
		return nil, errors.New("PipelineInfoGet failed")
	}

	return &pipeInfo, nil
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
