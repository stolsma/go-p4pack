// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package flowtest

import (
	"context"
	"errors"
	"fmt"

	"github.com/stolsma/go-p4pack/pkg/logging"
)

var flowTest *FlowTest
var log logging.Logger

func init() {
	// keep the logger up to date, also after new log config
	logging.Register("flowtest", func(logger logging.Logger) {
		log = logger
	})
}

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

// return a pointer to the (initialized) flowtest singleton
func Get() *FlowTest {
	return flowTest
}

// create and initialize the flowtest singleton, return the current flowtest singleton with error if it already exists
func CreateAndInit(ctx context.Context) (*FlowTest, error) {
	if flowTest != nil {
		return flowTest, errors.New("flowtest already initialized")
	}

	flowTest = &FlowTest{
		ifaces: make(map[string]*Iface),
	}

	flowTest.Init(ctx)

	return flowTest, nil
}

func (t *FlowTest) Init(ctx context.Context) {
	ctx, cancelFn := context.WithCancel(ctx)
	t.ctx = ctx
	t.cancelFn = cancelFn
}

func (t *FlowTest) AddInterface(ifName string, mac HexArray, ip HexArray) error {
	if t.ifaces[ifName] != nil {
		return fmt.Errorf("interface %s already exists", ifName)
	}

	t.ifaces[ifName] = &Iface{ // interface configuration
		Name: ifName,
		MAC:  mac,
		IP:   ip,
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
