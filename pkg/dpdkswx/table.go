// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package dpdkswx

/*
#cgo pkg-config: libdpdk

#include <rte_table_acl.h>
#include <rte_table_array.h>
#include <rte_table_hash.h>
#include <rte_table_lpm.h>
#include <rte_table_lpm_ipv6.h>

#include <rte_swx_pipeline.h>
#include <rte_swx_ctl.h>

*/
import "C"

type Table struct {
	id                   uint
	pipeline             *Pipeline
	name                 string // Table name.
	args                 string // Table creation arguments.
	nMatchFields         uint   // Number of match fields.
	nActions             uint   // Number of actions.
	defaultActionIsConst bool   // true => the default action is constant; false => the default action not constant
	size                 uint   // Table size parameter.
	clean                func() // the callback function called at clear
}

// Initialize table record after creation
func (t *Table) Init(p *Pipeline, id uint, clean func()) error {
	tableInfo, err := t.ctlTabelInfoGet(p, id)
	if err != nil {
		return err
	}

	t.id = id
	t.pipeline = p
	t.name = C.GoString(&tableInfo.name[0])
	t.args = C.GoString(&tableInfo.args[0])
	t.nMatchFields = (uint)(tableInfo.n_match_fields)
	t.nActions = (uint)(tableInfo.n_actions)
	if tableInfo.default_action_is_const == 0 {
		t.defaultActionIsConst = false
	} else {
		t.defaultActionIsConst = true
	}
	t.size = (uint)(tableInfo.size)
	t.clean = clean

	return nil
}

func (t *Table) Clear() {
	// TODO check if all memory related to this structure is freed
	// call given clean callback function if given during init
	if t.clean != nil {
		t.clean()
	}
}

func (t *Table) Name() string {
	return t.name
}

type CtlTableInfo C.struct_rte_swx_ctl_table_info

func (t *Table) ctlTabelInfoGet(p *Pipeline, id uint) (*CtlTableInfo, error) {
	var tableInfo CtlTableInfo

	res := (int)(C.rte_swx_ctl_table_info_get(p.Pipeline(), (C.uint32_t)(id), (*C.struct_rte_swx_ctl_table_info)(&tableInfo)))
	if res < 0 {
		return nil, err(res)
	}

	return &tableInfo, nil
}
