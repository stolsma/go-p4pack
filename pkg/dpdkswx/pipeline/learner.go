// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package pipeline

import (
	"fmt"
)

type LearnerTable struct {
	index                uint   // Index in swx_pipeline learnertable store
	name                 string // Learner Table name.
	nMatchFields         uint   // Number of match fields.
	nActions             uint   // Number of actions.
	defaultActionIsConst bool   // true => the default action is constant; false => the default action not constant
	size                 int    // Table size parameter.
	matchFields          TableMatchFieldStore
	actions              TableActionStore
}

// Initialize Learner table record from pipeline
func (t *LearnerTable) Init(p *Pipeline, index uint) error {
	learnerInfo, err := p.LearnerInfoGet(index)
	if err != nil {
		return err
	}

	// initalize generic table attributes
	t.index = index
	t.name = learnerInfo.GetName()
	t.nMatchFields = learnerInfo.GetNMatchFields()
	t.nActions = learnerInfo.GetNActions()
	t.defaultActionIsConst = learnerInfo.DefaultActionIsConst()
	t.size = int(learnerInfo.GetSize())

	// get all matchfields for this table
	t.matchFields = CreateTableMatchFieldsStore()
	for i := uint(0); i < t.nMatchFields; i++ {
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
	for i := uint(0); i < t.nActions; i++ {
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

func (t *LearnerTable) Clear() {
	// TODO check if all memory related to this structure is freed
	// call given clean callback function if given during init
}

func (t *LearnerTable) GetIndex() uint {
	return t.index
}

func (t *LearnerTable) GetName() string {
	return t.name
}

// true => the default action is constant; false => the default action not constant
func (t *LearnerTable) GetDefaultActionIsConst() bool {
	return t.defaultActionIsConst
}

func (t *LearnerTable) GetSize() int {
	return t.size
}

func (t *LearnerTable) GetMatchFields() TableMatchFieldStore {
	return t.matchFields
}

func (t *LearnerTable) GetActions() TableActionStore {
	return t.actions
}

// LearnerStore represents a store of LearnerTable records
type LearnerStore map[string]*LearnerTable

func CreateLearnerStore() LearnerStore {
	return make(LearnerStore)
}

func (lts LearnerStore) FindName(name string) *LearnerTable {
	if name == "" {
		return nil
	}

	return lts[name]
}

func (lts LearnerStore) CreateFromPipeline(p *Pipeline) error {
	pipelineInfo, err := p.PipelineInfoGet()
	if err != nil {
		return err
	}

	for i := uint(0); i < pipelineInfo.GetNLearners(); i++ {
		var ltable LearnerTable

		err := ltable.Init(p, i)
		if err != nil {
			return fmt.Errorf("LearnerStore.CreateFromPipeline error: %d", err)
		}
		lts.Add(&ltable)
	}

	return nil
}

func (lts LearnerStore) Add(ltable *LearnerTable) {
	lts[ltable.GetName()] = ltable
}

func (lts LearnerStore) ForEach(fn func(key string, table *LearnerTable) error) error {
	for k, v := range lts {
		if err := fn(k, v); err != nil {
			return err
		}
	}
	return nil
}

// Delete all LearnerTable records and free corresponding memory if required
func (lts LearnerStore) Clear() {
	for _, ltable := range lts {
		ltable.Clear()
		delete(lts, ltable.GetName())
	}
}
