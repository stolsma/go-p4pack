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
// @param ctl Pipeline control handle.
//
// @param action When non-zero (false), all the scheduled work is discarded after a failed commit. Otherwise, the
// scheduled work is still kept pending for the next commit. See CommitSaveOnFail and CommitAbortOnFail
//
// @return
// 0 on success or the following error codes otherwise:
//
//	-EINVAL: Invalid argument.
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
// @param tableName Table name.
//
// @param line String containing the table entry.
//
// @return *TableEntry on success or nil otherwise:
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
// @param tableName Table name.
//
// @param entry Entry to be added to the table.
//
// @return 0 on success or the following error codes otherwise:
// -EINVAL: Invalid argument.
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
// @param tableName Table name.
//
// @param entry The new table default entry. The *key* and *key_mask* entry fields are ignored.
//
// @return 0 on success or the following error codes otherwise:
// -EINVAL: Invalid argument.
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
// @param tableName Table name.
//
// @param entry Entry to be deleted from the table. The *action_id* and *action_data* entry fields are ignored.
//
// @return 0 on success or the following error codes otherwise:
// -EINVAL: Invalid argument.
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
// @param learnerName Learner table name.
//
// @param line String containing the learner table default entry.
//
// @return
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
// @param learnerName Learner table name.
//
// @param entry The new table default entry. The *key* and *key_mask* entry fields are ignored.
//
// @return 0 on success or the following error codes otherwise:
// -EINVAL: Invalid argument.
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
