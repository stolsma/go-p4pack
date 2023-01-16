// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package pipeline

import "fmt"

// represent an action argument description
type ActionArg struct {
	index         uint
	ActionArgInfo // TODO implement native Go struct instead of C struct
}

func (aa *ActionArg) GetIndex() uint {
	return aa.index
}

type ActionArgStore map[string]*ActionArg

type Action struct {
	index      uint           // index
	name       string         // action name.
	actionArgs ActionArgStore // action args
}

// Initialize Action record after creation. Returns nil on success or the following error codes otherwise:
//
//	-EINVAL = Invalid argument
func (a *Action) Init(p *Pipeline, index uint) error {
	actionInfo, err := p.ActionInfoGet(index)
	if err != nil {
		return err
	}

	// initalize generic table attibutes
	a.index = index
	a.name = actionInfo.GetName()
	a.actionArgs = make(ActionArgStore)

	// get action arg information
	for i := uint(0); i < actionInfo.GetNArgs(); i++ {
		actionArgInfo, err := p.ActionArgInfoGet(index, i)
		if err != nil {
			return err
		}

		var actionArg = &ActionArg{i, *actionArgInfo}
		a.actionArgs[actionArg.GetName()] = actionArg
	}

	return nil
}

func (a *Action) GetIndex() uint {
	return a.index
}

func (a *Action) GetName() string {
	return a.name
}

// represents a store of action records
type ActionStore map[string]*Action

func CreateActionStore() ActionStore {
	return make(ActionStore)
}

func (as ActionStore) FindName(name string) *Action {
	if name == "" {
		return nil
	}

	return as[name]
}

func (as ActionStore) FindIndex(index uint) *Action {
	for _, action := range as {
		if action.GetIndex() == index {
			return action
		}
	}
	return nil
}

func (as ActionStore) CreateFromPipeline(p *Pipeline) error {
	pipelineInfo, err := p.PipelineInfoGet()
	if err != nil {
		return err
	}

	for i := uint(0); i < pipelineInfo.GetNActions(); i++ {
		var action Action

		err := action.Init(p, i)
		if err != nil {
			return fmt.Errorf("Actionstore.CreateFromPipeline error: %d", err)
		}
		as.Add(&action)
	}

	return nil
}

func (as ActionStore) Add(action *Action) {
	as[action.GetName()] = action
}

// Delete all Action records and free corresponding memory if required
func (as ActionStore) Clear() {
	for _, action := range as {
		delete(as, action.GetName())
	}
}
