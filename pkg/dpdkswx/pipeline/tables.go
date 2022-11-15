// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package pipeline

import (
	"fmt"
)

type TableMatchField struct {
	index     int
	matchType int
	isHeader  bool
	nBits     int
	offset    int
}

func (tmf *TableMatchField) GetIndex() int {
	return tmf.index
}

// TableMatchFieldsStore represents a store of TableMatchFields records
type TableMatchFieldStore map[int]*TableMatchField

func CreateTableMatchFieldsStore() TableMatchFieldStore {
	return make(TableMatchFieldStore)
}

func (tmfs TableMatchFieldStore) FindIndex(index int) *TableMatchField {
	for _, tableMatchField := range tmfs {
		if tableMatchField.GetIndex() == index {
			return tableMatchField
		}
	}
	return nil
}

func (tmfs TableMatchFieldStore) Add(tableMatchField *TableMatchField) {
	tmfs[tableMatchField.GetIndex()] = tableMatchField
}

func (tmfs TableMatchFieldStore) ForEach(fn func(key int, tableMatchField *TableMatchField) error) error {
	for k, v := range tmfs {
		if err := fn(k, v); err != nil {
			return err
		}
	}
	return nil
}

// Delete all TableMatchField records and free corresponding memory if required
func (tmfs TableMatchFieldStore) Clear() {
	for _, tableMatchField := range tmfs {
		delete(tmfs, tableMatchField.GetIndex())
	}
}

type TableAction struct {
	index                   int
	action                  *Action
	actionIsForDefaultEntry bool
	actionIsForTableEntries bool
}

func (ta *TableAction) GetIndex() int {
	return ta.index
}

func (ta *TableAction) GetActionName() string {
	return ta.action.GetName()
}

func (ta *TableAction) GetActionIndex() int {
	return ta.action.GetIndex()
}

func (ta *TableAction) GetActionIsForDefaultEntry() bool {
	return ta.actionIsForDefaultEntry
}

func (ta *TableAction) GetActionIsForTableEntries() bool {
	return ta.actionIsForTableEntries
}

// TableActionStore represents a store of TableAction records
type TableActionStore map[string]*TableAction

func CreateTableActionStore() TableActionStore {
	return make(TableActionStore)
}

func (tas TableActionStore) FindName(name string) *TableAction {
	if name == "" {
		return nil
	}

	return tas[name]
}

func (tas TableActionStore) FindIndex(index int) *TableAction {
	for _, tableAction := range tas {
		if tableAction.GetIndex() == index {
			return tableAction
		}
	}
	return nil
}

func (tas TableActionStore) Add(tableAction *TableAction) {
	tas[tableAction.GetActionName()] = tableAction
}

func (tas TableActionStore) ForEach(fn func(key string, tableAction *TableAction) error) error {
	for k, v := range tas {
		if err := fn(k, v); err != nil {
			return err
		}
	}
	return nil
}

// Delete all TableAction records and free corresponding memory if required
func (tas TableActionStore) Clear() {
	for _, tableAction := range tas {
		delete(tas, tableAction.GetActionName())
	}
}

type Table struct {
	index                int    // Index in swx_pipeline table store
	name                 string // Table name.
	args                 string // Table creation arguments.
	nMatchFields         int    // Number of match fields.
	nActions             int    // Number of actions.
	defaultActionIsConst bool   // true => the default action is constant; false => the default action not constant
	size                 int    // Table size parameter.
	matchFields          TableMatchFieldStore
	actions              TableActionStore
}

// Initialize table record from pipeline
func (t *Table) Init(p *Pipeline, index int) error {
	tableInfo, err := p.TableInfoGet(index)
	if err != nil {
		return err
	}

	// initalize generic table attributes
	t.index = index
	t.name = tableInfo.GetName()
	t.args = tableInfo.GetArgs()
	t.nMatchFields = tableInfo.GetNMatchFields()
	t.nActions = tableInfo.GetNActions()
	t.defaultActionIsConst = tableInfo.GetDefaultActionIsConst()
	t.size = tableInfo.GetSize()

	// get all matchfields for this table
	t.matchFields = CreateTableMatchFieldsStore()
	for i := 0; i < t.nMatchFields; i++ {
		tableMatchFieldInfo, err := p.TableMatchFieldInfoGet(index, i)
		if err != nil {
			return err
		}

		tableMatchField := TableMatchField{
			index:     i,
			matchType: tableMatchFieldInfo.GetMatchType(),
			isHeader:  tableMatchFieldInfo.GetIsHeader(),
			nBits:     tableMatchFieldInfo.GetNBits(),
			offset:    tableMatchFieldInfo.GetOffset(),
		}
		t.matchFields.Add(&tableMatchField)
	}

	// get all actions for this table
	t.actions = CreateTableActionStore()
	for i := 0; i < t.nActions; i++ {
		tableActionInfo, err := p.TableActionInfoGet(index, i)
		if err != nil {
			return err
		}

		action := p.actions.FindIndex(tableActionInfo.GetActionID())
		if action == nil {
			return fmt.Errorf("didn't find TableAction (ActionID: %d) for table %s",
				tableActionInfo.GetActionID(), t.GetName())
		}

		tableAction := TableAction{
			index:                   i,
			action:                  action,
			actionIsForDefaultEntry: tableActionInfo.GetActionIsForDefaultEntry(),
			actionIsForTableEntries: tableActionInfo.GetActionIsForTableEntries(),
		}
		t.actions.Add(&tableAction)
	}

	return nil
}

func (t *Table) Clear() {
	// TODO check if all memory related to this structure is freed
	// call given clean callback function if given during init
}

func (t *Table) GetIndex() int {
	return t.index
}

func (t *Table) GetName() string {
	return t.name
}

func (t *Table) GetArgs() string {
	return t.args
}

// true => the default action is constant; false => the default action not constant
func (t *Table) GetDefaultActionIsConst() bool {
	return t.defaultActionIsConst
}

func (t *Table) GetSize() int {
	return t.size
}

func (t *Table) GetMatchFields() TableMatchFieldStore {
	return t.matchFields
}

func (t *Table) GetActions() TableActionStore {
	return t.actions
}

// TableStore represents a store of Table records
type TableStore map[string]*Table

func CreateTableStore() TableStore {
	return make(TableStore)
}

func (ts TableStore) FindName(name string) *Table {
	if name == "" {
		return nil
	}

	return ts[name]
}

func (ts TableStore) CreateFromPipeline(p *Pipeline) error {
	pipelineInfo, err := p.PipelineInfoGet()
	if err != nil {
		return err
	}

	for i := 0; i < pipelineInfo.GetNTables(); i++ {
		var table Table

		err := table.Init(p, i)
		if err != nil {
			return fmt.Errorf("Tablestore.CreateFromPipeline error: %d", err)
		}
		ts.Add(&table)
	}

	return nil
}

func (ts TableStore) Add(table *Table) {
	ts[table.GetName()] = table
}

func (ts TableStore) ForEach(fn func(key string, table *Table) error) error {
	for k, v := range ts {
		if err := fn(k, v); err != nil {
			return err
		}
	}
	return nil
}

// Delete all Table records and free corresponding memory if required
func (ts TableStore) Clear() {
	for _, table := range ts {
		table.Clear()
		delete(ts, table.GetName())
	}
}
