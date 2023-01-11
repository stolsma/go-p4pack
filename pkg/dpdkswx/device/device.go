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

// Methods all port devices must implement
type Type interface {
	Type() string
	SetType(devType string)
	Name() string
	SetName(name string)
	InitializeQueues(nRxQ uint16, nTxQ uint16)
	IterateRxQueues(fn func(index uint16, q Queue) error) error
	IterateTxQueues(fn func(index uint16, q Queue) error) error
	GetRxQueue(q uint16) (string, int, error)
	SetRxQueue(q uint16, pl string, portID int) error
	GetTxQueue(q uint16) (string, int, error)
	SetTxQueue(q uint16, pl string, portID int) error
	Clean() func()
	SetClean(fn func())
	BindToPipelineInputPort(pl *pipeline.Pipeline, portID int, rxq uint16, bsz uint) error
	BindToPipelineOutputPort(pl *pipeline.Pipeline, portID int, txq uint16, bsz uint) error
	IsBound() bool
	SetLinkUp() error
	SetLinkDown() error
	GetPortInfo() (map[string]string, error)
	GetPortStats() (map[string]string, error)
}

type Queue struct {
	pipeline     string
	pipelinePort int
}

func (q *Queue) Pipeline() string {
	return q.pipeline
}

func (q *Queue) PipelinePort() int {
	return q.pipelinePort
}

// basic definition all port devices need to "inherit"
type Device struct {
	devType  string
	name     string
	nRxQ     uint16
	rxQueues []Queue
	nTxQ     uint16
	txQueues []Queue
	clean    func()
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

func (d *Device) InitializeQueues(nRxQ uint16, nTxQ uint16) {
	d.nRxQ = nRxQ
	for i := uint16(0); i < nRxQ; i++ {
		d.rxQueues = append(d.rxQueues, Queue{"", NotBound})
	}

	d.nTxQ = nTxQ
	for i := uint16(0); i < nTxQ; i++ {
		d.txQueues = append(d.txQueues, Queue{"", NotBound})
	}
}

func (d *Device) IterateRxQueues(fn func(index uint16, q Queue) error) error {
	if fn != nil && d.nRxQ > 0 {
		for i := uint16(0); i < d.nRxQ; i++ {
			if err := fn(i, d.rxQueues[i]); err != nil {
				return err
			}
		}
	} else {
		return errors.New("no function to call or configured queues is 0")
	}
	return nil
}

func (d *Device) IterateTxQueues(fn func(index uint16, q Queue) error) error {
	if fn != nil && d.nTxQ > 0 {
		for i := uint16(0); i < d.nTxQ; i++ {
			if err := fn(i, d.txQueues[i]); err != nil {
				return err
			}
		}
	} else {
		return errors.New("no function to call or configured queues is 0")
	}
	return nil
}

func (d *Device) GetRxQueue(q uint16) (string, int, error) {
	if q >= d.nRxQ {
		return "", -1, errors.New("requested queue not available")
	}
	return d.rxQueues[q].pipeline, d.rxQueues[q].pipelinePort, nil
}

func (d *Device) SetRxQueue(q uint16, pl string, portID int) error {
	if q >= d.nRxQ {
		return errors.New("requested queue not available")
	}
	d.rxQueues[q].pipeline = pl
	d.rxQueues[q].pipelinePort = portID
	return nil
}

func (d *Device) GetTxQueue(q uint16) (string, int, error) {
	if q >= d.nTxQ {
		return "", -1, errors.New("requested queue not available")
	}
	return d.txQueues[q].pipeline, d.txQueues[q].pipelinePort, nil
}

func (d *Device) SetTxQueue(q uint16, pl string, portID int) error {
	if q >= d.nTxQ {
		return errors.New("requested queue not available")
	}
	d.txQueues[q].pipeline = pl
	d.txQueues[q].pipelinePort = portID
	return nil
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

// return true if one of the queues (rx and tx) is bound to a pipeline
func (d *Device) IsBound() bool {
	if err := d.IterateRxQueues(func(index uint16, q Queue) error {
		if q.pipelinePort != NotBound {
			return errors.New("this queue is bound")
		}
		return nil
	}); err != nil {
		return true
	}

	if err := d.IterateTxQueues(func(index uint16, q Queue) error {
		if q.pipelinePort != NotBound {
			return errors.New("this queue is bound")
		}
		return nil
	}); err != nil {
		return true
	}

	return false
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
