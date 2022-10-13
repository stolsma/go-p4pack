// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package flowtest

import (
	"context"
)

type Iface struct {
	Name string
	MAC  HexArray
	IP   HexArray
}

func (i *Iface) GetName() string {
	return i.Name
}

func (i *Iface) GetMAC() HexArray {
	return i.MAC
}

func (i *Iface) GetIP() HexArray {
	return i.IP
}

type IfaceMap map[string]*Iface

type FlowTest struct {
	ctx      context.Context
	cancelFn context.CancelFunc
	ifaces   map[string]*Iface
	flowSets []*FlowSet
}

func Create() *FlowTest {
	t := &FlowTest{
		ifaces: make(map[string]*Iface),
	}

	return t
}

func Init(ctx context.Context, config Config) (*FlowTest, error) {
	t := Create()
	return t, t.Init(ctx, config)
}

func (t *FlowTest) Init(ctx context.Context, config Config) error {
	ctx, cancelFn := context.WithCancel(ctx)
	t.ctx = ctx
	t.cancelFn = cancelFn

	// initialize the interface references
	for _, intf := range config.Interfaces {
		ifName := intf.GetName()
		t.ifaces[ifName] = &Iface{ // interface configuration
			Name: ifName,
			MAC:  intf.GetMAC(),
			IP:   intf.GetIP(),
		}
	}

	for _, test := range config.FlowSets {
		err := t.AddFlowSet(test)
		if err != nil {
			return err
		}
	}

	return nil
}

func (t *FlowTest) AddFlowSet(config FlowSetConfig) error {
	fs := FlowSetCreate(config.GetName())

	err := fs.Init(config.Flows, t.ifaces)
	if err != nil {
		return err
	}

	t.flowSets = append(t.flowSets, fs)

	return nil
}

func (t *FlowTest) StartAll() error {
	for _, fs := range t.flowSets {
		err := fs.Start(t.ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

func (t *FlowTest) StopAll() error {
	for _, fs := range t.flowSets {
		err := fs.Stop()
		if err != nil {
			return err
		}
	}
	return nil
}
