// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package device

import (
	"errors"

	"github.com/stolsma/go-p4pack/pkg/dpdkswx/pipeline"
)

const (
	NotBound int = -1
)

var ErrNotImplemented = errors.New("not implemented")

type Type interface {
	Type() string
	SetType(devType string)
	Name() string
	SetName(name string)
	PipelineIn() string
	SetPipelineIn(pl string)
	PipelineInPort() int
	SetPipelineInPort(portID int)
	PipelineOut() string
	SetPipelineOut(pl string)
	PipelineOutPort() int
	SetPipelineOutPort(portID int)
	Clean() func()
	SetClean(fn func())
	BindToPipelineInputPort(pl *pipeline.Pipeline, portID int, rxq uint, bsz uint) error
	BindToPipelineOutputPort(pl *pipeline.Pipeline, portID int, txq uint, bsz uint) error
	SetLinkUp() error
	SetLinkDown() error
	GetPortInfo() (map[string]string, error)
	GetPortStats() (map[string]string, error)
}

type Device struct {
	devType         string
	name            string
	pipelineIn      string
	pipelineInPort  int
	pipelineOut     string
	pipelineOutPort int
	clean           func()
}

func (d *Device) Type() string {
	return d.devType
}

func (d *Device) SetType(devType string) {
	d.devType = devType
}

func (d *Device) Name() string {
	return d.name
}

func (d *Device) SetName(name string) {
	d.name = name
}

func (d *Device) PipelineIn() string {
	return d.pipelineIn
}

func (d *Device) SetPipelineIn(pl string) {
	d.pipelineIn = pl
}

func (d *Device) PipelineInPort() int {
	return d.pipelineInPort
}

func (d *Device) SetPipelineInPort(portID int) {
	d.pipelineInPort = portID
}

func (d *Device) PipelineOut() string {
	return d.pipelineOut
}

func (d *Device) SetPipelineOut(pl string) {
	d.pipelineOut = pl
}

func (d *Device) PipelineOutPort() int {
	return d.pipelineOutPort
}

func (d *Device) SetPipelineOutPort(portID int) {
	d.pipelineOutPort = portID
}

func (d *Device) Clean() func() {
	return d.clean
}

func (d *Device) SetClean(fn func()) {
	d.clean = fn
}

func (d *Device) BindToPipelineInputPort(pl *pipeline.Pipeline, portID int, rxq uint, bsz uint) error {
	return ErrNotImplemented
}

// bind to given pipeline output port
func (d *Device) BindToPipelineOutputPort(pl *pipeline.Pipeline, portID int, txq uint, bsz uint) error {
	return ErrNotImplemented
}

func (d *Device) SetLinkUp() error {
	return ErrNotImplemented
}

func (d *Device) SetLinkDown() error {
	return ErrNotImplemented
}

func (d *Device) GetPortInfo() (map[string]string, error) {
	return map[string]string{}, ErrNotImplemented
}

func (d *Device) GetPortStats() (map[string]string, error) {
	return map[string]string{}, ErrNotImplemented
}
