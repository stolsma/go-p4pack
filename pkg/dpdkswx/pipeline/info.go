// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package pipeline

/*
#include <rte_swx_pipeline.h>
#include <rte_swx_ctl.h>
*/
import "C"
import (
	"fmt"

	"github.com/stolsma/go-p4pack/pkg/dpdkswx/common"
)

type Info C.struct_rte_swx_ctl_pipeline_info

func (pi *Info) GetNPortsIn() uint {
	return uint(pi.n_ports_in)
}

func (pi *Info) GetNPortsOut() uint {
	return uint(pi.n_ports_out)
}

func (pi *Info) GetNMirroringSlots() uint {
	return uint(pi.n_mirroring_slots)
}

func (pi *Info) GetNMirroringSessions() uint {
	return uint(pi.n_mirroring_sessions)
}

func (pi *Info) GetNActions() uint {
	return uint(pi.n_actions)
}

func (pi *Info) GetNTables() uint {
	return uint(pi.n_tables)
}

func (pi *Info) GetNSelectors() uint {
	return uint(pi.n_selectors)
}

func (pi *Info) GetNLearners() uint {
	return uint(pi.n_learners)
}

func (pi *Info) GetNRegarrays() uint {
	return uint(pi.n_regarrays)
}

func (pi *Info) GetNMetarrays() uint {
	return uint(pi.n_metarrays)
}

const infoTemplate = `ports in           : %d
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

// Multiline pipeline info string
func (pi *Info) String() string {
	return fmt.Sprintf(infoTemplate, pi.n_ports_in, pi.n_ports_out, pi.n_mirroring_slots, pi.n_mirroring_sessions,
		pi.n_actions, pi.n_tables, pi.n_selectors, pi.n_learners, pi.n_regarrays, pi.n_metarrays)
}

// Pipeline info get
//
// Get the pipeline info. Returns Info on success or the following error codes otherwise:
//
//	-EINVAL = Invalid argument
func (pl *Pipeline) PipelineInfoGet() (*Info, error) {
	var pipeInfo Info

	if status := C.rte_swx_ctl_pipeline_info_get(pl.p, (*C.struct_rte_swx_ctl_pipeline_info)(&pipeInfo)); status != 0 {
		return nil, common.Err(status)
	}
	return &pipeInfo, nil
}

type ActionInfo C.struct_rte_swx_ctl_action_info

func (ai *ActionInfo) GetName() string {
	return C.GoString(&ai.name[0])
}

func (ai *ActionInfo) GetNArgs() uint {
	return (uint)(ai.n_args)
}

// Action info get
//
// Get the action (actionID) info. Returns ActionInfo on success or the following error codes otherwise:
//
//	-EINVAL = Invalid argument
func (pl *Pipeline) ActionInfoGet(actionID uint) (*ActionInfo, error) {
	var actionInfo = &ActionInfo{}

	if status := C.rte_swx_ctl_action_info_get(pl.p, (C.uint)(actionID),
		(*C.struct_rte_swx_ctl_action_info)(actionInfo),
	); status != 0 {
		return nil, common.Err(status)
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

// Action arguments info get
//
// Get the action (actionID) argument (actionArgID) info. Returns ActionArgInfo on success or the following error codes
// otherwise:
//
//	-EINVAL = Invalid argument
func (pl *Pipeline) ActionArgInfoGet(actionID uint, actionArgID uint) (*ActionArgInfo, error) {
	var actionArgInfo = &ActionArgInfo{}

	if status := C.rte_swx_ctl_action_arg_info_get(pl.p, (C.uint)(actionID), (C.uint)(actionArgID),
		(*C.struct_rte_swx_ctl_action_arg_info)(actionArgInfo),
	); status != 0 {
		return nil, common.Err(status)
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

func (ti *TableInfo) GetNMatchFields() uint {
	return uint(ti.n_match_fields)
}

func (ti *TableInfo) GetNActions() uint {
	return uint(ti.n_actions)
}

func (ti *TableInfo) GetDefaultActionIsConst() bool {
	return ti.default_action_is_const > 0
}

func (ti *TableInfo) GetSize() int {
	return int(ti.size)
}

// Table info get
//
// Get the table (tableID) info. Returns TableInfo on success or the following error codes otherwise:
//
//	-EINVAL = Invalid argument
func (pl *Pipeline) TableInfoGet(tableID uint) (*TableInfo, error) {
	var tableInfo TableInfo

	if status := C.rte_swx_ctl_table_info_get(pl.p, (C.uint)(tableID), (*C.struct_rte_swx_ctl_table_info)(&tableInfo)); status != 0 {
		return nil, common.Err(status)
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

// Table match field info get
//
// Get the table (tableID) match field (matchFieldID) info. Returns TableMatchFieldInfo on success or the following
// error codes otherwise:
//
//	-EINVAL = Invalid argument
func (pl *Pipeline) TableMatchFieldInfoGet(tableID uint, matchFieldID uint) (*TableMatchFieldInfo, error) {
	var tableMatchFieldInfo TableMatchFieldInfo

	if status := C.rte_swx_ctl_table_match_field_info_get(
		pl.p, (C.uint)(tableID), (C.uint)(matchFieldID), (*C.struct_rte_swx_ctl_table_match_field_info)(&tableMatchFieldInfo),
	); status != 0 {
		return nil, common.Err(status)
	}
	return &tableMatchFieldInfo, nil
}

// information about table actions
type TableActionInfo C.struct_rte_swx_ctl_table_action_info

func (tai *TableActionInfo) GetActionID() uint {
	return uint(tai.action_id)
}

func (tai *TableActionInfo) GetActionIsForTableEntries() bool {
	return tai.action_is_for_table_entries > 0
}

func (tai *TableActionInfo) GetActionIsForDefaultEntry() bool {
	return tai.action_is_for_default_entry > 0
}

// Table action info get
//
// Get the table (tableID) action (actionID) info. Returns TableActionInfo on success or the following error codes
// otherwise:
//
//	-EINVAL = Invalid argument
func (pl *Pipeline) TableActionInfoGet(tableID uint, actionID uint) (*TableActionInfo, error) {
	var tableActionInfo TableActionInfo

	if status := C.rte_swx_ctl_table_action_info_get(
		pl.p, (C.uint)(tableID), (C.uint)(actionID), (*C.struct_rte_swx_ctl_table_action_info)(&tableActionInfo),
	); status != 0 {
		return nil, common.Err(status)
	}
	return &tableActionInfo, nil
}

// information about the structure of a learner table
type LearnerInfo C.struct_rte_swx_ctl_learner_info

// return the name of the learner table
func (li *LearnerInfo) GetName() string {
	return C.GoString(&li.name[0])
}

func (li *LearnerInfo) GetNMatchFields() uint {
	return uint(li.n_match_fields)
}

func (li *LearnerInfo) GetNActions() uint {
	return uint(li.n_actions)
}

func (li *LearnerInfo) DefaultActionIsConst() bool {
	return li.default_action_is_const > 0
}

func (li *LearnerInfo) GetSize() uint32 {
	return uint32(li.size)
}

func (li *LearnerInfo) GetNKeyTimeouts() uint32 {
	return uint32(li.n_key_timeouts)
}

// Learner info get
//
// Get the learner table (tableID) info. Returns LearnerInfo on success or the following error codes otherwise:
//
//	-EINVAL = Invalid argument
func (pl *Pipeline) LearnerInfoGet(tableID uint) (*LearnerInfo, error) {
	var learnerInfo LearnerInfo

	if status := C.rte_swx_ctl_learner_info_get(
		pl.p, (C.uint)(tableID), (*C.struct_rte_swx_ctl_learner_info)(&learnerInfo),
	); status != 0 {
		return nil, common.Err(status)
	}
	return &learnerInfo, nil
}

// information about the structure of a register array
type RegarrayInfo C.struct_rte_swx_ctl_regarray_info

func (ra *RegarrayInfo) GetName() string {
	return C.GoString(&ra.name[0])
}

func (ra *RegarrayInfo) GetSize() int {
	return int(ra.size)
}

// Register array info get
//
// Get the register array (regarrayID) info. Returns RegarrayInfo on success or the following error codes otherwise:
//
//	-EINVAL = Invalid argument
func (pl *Pipeline) RegarrayInfoGet(regarrayID uint) (*RegarrayInfo, error) {
	var regarrayInfo RegarrayInfo

	if status := C.rte_swx_ctl_regarray_info_get(
		pl.p, (C.uint)(regarrayID), (*C.struct_rte_swx_ctl_regarray_info)(&regarrayInfo),
	); status != 0 {
		return nil, common.Err(status)
	}
	return &regarrayInfo, nil
}

// information about the structure of a meter array
type MetarrayInfo C.struct_rte_swx_ctl_metarray_info

func (ma *MetarrayInfo) GetName() string {
	return C.GoString(&ma.name[0])
}

func (ma *MetarrayInfo) GetSize() int {
	return int(ma.size)
}

// Meter array info get
//
// Get the meter array (metarrayID) info. Returns MetarrayInfo on success or the following error codes otherwise:
//
//	-EINVAL = Invalid argument
func (pl *Pipeline) MetarrayInfoGet(metarrayID uint) (*MetarrayInfo, error) {
	var metarrayInfo MetarrayInfo

	if status := C.rte_swx_ctl_metarray_info_get(
		pl.p, (C.uint)(metarrayID), (*C.struct_rte_swx_ctl_metarray_info)(&metarrayInfo),
	); status != 0 {
		return nil, common.Err(status)
	}
	return &metarrayInfo, nil
}
