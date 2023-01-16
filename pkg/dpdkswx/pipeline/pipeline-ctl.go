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

#include <rte_swx_pipeline.h>
#include <rte_swx_ctl.h>
*/
import "C"
import (
	"errors"
	"fmt"
	"unsafe"

	"github.com/stolsma/go-p4pack/pkg/dpdkswx/common"
)

// The Pipeline control handling structure
type Ctl struct {
	ctl *C.struct_rte_swx_ctl_pipeline // Struct definition only swx internal
}

func CreatePipelineCtl(pl *Pipeline) (*Ctl, error) {
	pctl := &Ctl{}
	return pctl, pctl.Init(pl)
}

func (pctl *Ctl) Init(pl *Pipeline) error {
	pctl.ctl = C.rte_swx_ctl_pipeline_create((*C.struct_rte_swx_pipeline)(pl.GetPipeline()))
	if pctl.ctl == nil {
		return errors.New("rte_swx_ctl_pipeline_create error")
	}

	return nil
}

// Pipeline control struct free. If internal ctl struct pointer is nil, no operation is performed.
func (pctl *Ctl) Free() {
	if pctl.ctl == nil {
		return
	}

	C.rte_swx_ctl_pipeline_free(pctl.ctl)
	pctl.ctl = nil
}

// Commit Action on fail type
type CommitAction int

const (
	// Abort pipeline commit on fail, but keep the scheduled work pending for the next commit
	CommitSaveOnFail  CommitAction = 0 + iota
	CommitAbortOnFail              // Abort pipeline commit on fail and all the scheduled work is discarded
)

// Execute all the scheduled pipeline table work.
//
// When action is CommitAbortOnFail, all the scheduled work is discarded after a failed commit. Otherwise, with
// CommitSaveOnFail scheduled work is still kept pending for the next commit. See [pipeline.CommitSaveOnFail] and
// [pipeline.CommitAbortOnFail]. Returns 0 on success or the following error codes otherwise:
//
//	-EINVAL = Invalid argument.
func (pctl *Ctl) Commit(action CommitAction) error {
	res := C.rte_swx_ctl_pipeline_commit(pctl.ctl, (C.int)(action))
	return common.Err(res)
}

// Discard all the scheduled pipeline table work.
func (pctl *Ctl) Abort() {
	C.rte_swx_ctl_pipeline_abort(pctl.ctl)
}

type TableEntry C.struct_rte_swx_table_entry

// Free the memory occupied by the table entry C struct
func (te *TableEntry) Free() {
	if te == nil {
		return
	}

	C.free(unsafe.Pointer(te.key))
	C.free(unsafe.Pointer(te.key_mask))
	C.free(unsafe.Pointer(te.action_data))
	C.free(unsafe.Pointer(te))
}

// Read table entry definition from string and create TableEntry struct.
//
// The tableName argument contains the name of the table to create a TableEntry for represented by the line
// string. The line string containing the table entry. Returns a pointer to a filled TableEntry or nil
// if something is not correct with the given line string or if it is a comment string.
func (pctl *Ctl) TableEntryRead(tableName string, line string) *TableEntry {
	var isBlankOrComment C.int

	cTableName := C.CString(tableName)
	defer C.free(unsafe.Pointer(cTableName))
	cLine := C.CString(line)
	defer C.free(unsafe.Pointer(cLine))

	entry := (*TableEntry)(C.rte_swx_ctl_pipeline_table_entry_read(pctl.ctl, cTableName, cLine, &isBlankOrComment))

	if isBlankOrComment != 0 {
		return nil
	}

	return entry
}

// Schedule entry for addition to table or update as part of the next commit operation.
//
// The tableName argument contains the name of the table to add the new entry to. Returns 0 on success or the following
// error codes otherwise:
//
//	-EINVAL = Invalid argument.
func (pctl *Ctl) TableEntryAdd(tableName string, entry *TableEntry) error {
	cTableName := C.CString(tableName)
	defer C.free(unsafe.Pointer(cTableName))

	status := C.rte_swx_ctl_pipeline_table_entry_add(pctl.ctl, cTableName, (*C.struct_rte_swx_table_entry)(entry))
	entry.Free()

	if status != 0 {
		return fmt.Errorf("entry add error: %d", status)
	}

	return nil
}

// Schedule table default entry update as part of the next commit operation.
//
// The tableName argument contains the name of the table to add the new table default entry to. The *key* and *key_mask*
// entry fields are ignored. Returns 0 on success or the following error codes otherwise:
//
//	-EINVAL = Invalid argument.
func (pctl *Ctl) TableDefaultEntryAdd(tableName string, entry *TableEntry) error {
	cTableName := C.CString(tableName)
	defer C.free(unsafe.Pointer(cTableName))

	status := C.rte_swx_ctl_pipeline_table_default_entry_add(pctl.ctl, cTableName, (*C.struct_rte_swx_table_entry)(entry))
	entry.Free()

	if status != 0 {
		return fmt.Errorf("entry add error: %d", status)
	}

	return nil
}

// Schedule entry for deletion from table as part of the next commit operation. Request is silently discarded if no
// such entry exists.
//
// The tableName argument contains the name of the table to delete the entry from. The *action_id* and *action_data*
// entry fields are ignored. Returns 0 on success or the following error codes otherwise:
//
//	-EINVAL = Invalid argument.
func (pctl *Ctl) TableEntryDelete(tableName string, entry *TableEntry) error {
	cTableName := C.CString(tableName)
	defer C.free(unsafe.Pointer(cTableName))

	status := C.rte_swx_ctl_pipeline_table_entry_delete(pctl.ctl, cTableName, (*C.struct_rte_swx_table_entry)(entry))
	entry.Free()

	if status != 0 {
		return fmt.Errorf("entry delete error: %d", status)
	}

	return nil
}

// Read learner table default entry from string.
//
// The learnerName argument contains the name of the learner table to create a TableEntry for represented by the line
// string. The line string containing the learner table default entry. Returns a pointer to a filled TableEntry or nil
// if something is not correct with the given line string or if it is a comment string.
func (pctl *Ctl) LearnerDefaultEntryRead(learnerName string, line string) *TableEntry {
	var isBlankOrComment C.int

	cLearnerName := C.CString(learnerName)
	defer C.free(unsafe.Pointer(cLearnerName))
	cLine := C.CString(line)
	defer C.free(unsafe.Pointer(cLine))

	entry := (*TableEntry)(C.rte_swx_ctl_pipeline_learner_default_entry_read(pctl.ctl, cLearnerName, cLine, &isBlankOrComment))

	if isBlankOrComment != 0 {
		return nil
	}

	return entry
}

// Schedule learner table default entry update as part of the next commit operation.
//
// The learnerName argument contains the name of the learner table to add a new table default entry to. The *key* and
// *key_mask* entry fields are ignored. Returns nil on success or the following error codes otherwise:
//
//	-EINVAL = Invalid argument.
func (pctl *Ctl) LearnerDefaultEntryAdd(learnerName string, entry *TableEntry) error {
	cLearnerName := C.CString(learnerName)
	defer C.free(unsafe.Pointer(cLearnerName))

	status := C.rte_swx_ctl_pipeline_learner_default_entry_add(pctl.ctl, cLearnerName, (*C.struct_rte_swx_table_entry)(entry))
	entry.Free()

	if status != 0 {
		return fmt.Errorf("entry add error: %d", status)
	}

	return nil
}

// Pipeline selector table group add
//
// Add a new selector table group to a selector table (selector). This operation is executed before this function
// returns and its result is independent of the result of the next commit operation. Returns the the ID of the new
// group, which is only valid when the function call is successful. This group is initially empty, i.e. it does not
// contain any members. error is nil on success or the following error codes otherwise:
//
//	-EINVAL = Invalid argument
//	-ENOSPC = All groups are currently in use, no group available
func (pctl *Ctl) SelectorGroupAdd(selector string) (uint32, error) {
	var groupID C.uint
	cSelector := C.CString(selector)
	defer C.free(unsafe.Pointer(cSelector))

	if status := C.rte_swx_ctl_pipeline_selector_group_add(pctl.ctl, cSelector, &groupID); status != 0 {
		return 0, common.Err(status)
	}

	return uint32(groupID), nil
}

// Pipeline selector table group delete
//
// Schedule a selector table (selector) group (groupID) for deletion as part of the next commit operation. The group to
// be deleted can be empty or non-empty. Returns nil on success or the following error codes otherwise:
//
//	-EINVAL = Invalid argument
//	-ENOMEM = Not enough memory
func (pctl *Ctl) SelectorGroupDelete(selector string, groupID uint32) error {
	cSelector := C.CString(selector)
	defer C.free(unsafe.Pointer(cSelector))

	if status := C.rte_swx_ctl_pipeline_selector_group_delete(pctl.ctl, cSelector, C.uint(groupID)); status != 0 {
		return common.Err(status)
	}

	return nil
}

// Selector table member add to group
//
// Schedule the operation to add a new member (memberID) to an existing selector table (selector) group (groupID) as
// part of the next commit operation. If this member is already in this group, the member weight is updated to the new
// value. A weight of zero means this member is to be deleted from the group. Returns nil on success or the following
// error codes otherwise:
//
//	-EINVAL = Invalid argument
//	-ENOMEM = Not enough memory
//	-ENOSPC = The group is full
func (pctl *Ctl) SelectorGroupMemberAdd(selector string, groupID uint32, memberID uint32, memberWeight uint32) error {
	cSelector := C.CString(selector)
	defer C.free(unsafe.Pointer(cSelector))

	if status := C.rte_swx_ctl_pipeline_selector_group_member_add(pctl.ctl, cSelector, C.uint(groupID),
		C.uint(memberID), C.uint(memberWeight),
	); status != 0 {
		return common.Err(status)
	}

	return nil
}

// Selector table member delete from group
//
// Schedule the operation to delete a member (memberID) from an existing selector table (selector) group (groupID) as
// part of the next commit operation. Returns nil on success or the following error codes otherwise:
//
//	-EINVAL = Invalid argument
func (pctl *Ctl) SelectorGroupMemberDelete(selector string, groupID uint32, memberID uint32) error {
	cSelector := C.CString(selector)
	defer C.free(unsafe.Pointer(cSelector))

	if status := C.rte_swx_ctl_pipeline_selector_group_member_delete(pctl.ctl, cSelector, C.uint(groupID),
		C.uint(memberID),
	); status != 0 {
		return common.Err(status)
	}

	return nil
}
