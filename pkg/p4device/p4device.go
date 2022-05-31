// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package p4device

import (
	"fmt"
	"net"

	"github.com/stolsma/go-p4dpdk-vswitch/pkg/gnmi"
	"github.com/stolsma/go-p4dpdk-vswitch/pkg/p4rt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const GNMIPort = "9339"
const P4RTPort = "9559"

// P4Device defines a DPDK SWX target device instance
type P4Device struct {
	addr           string
	stop           func()
	gnmiGrpcServer *grpc.Server
	gnmiServer     *gnmi.Server
	p4rtGrpcServer *grpc.Server
	p4rtServer     *p4rt.Server
}

// New returns a new initialized P4Device instance.
func New(addr string, opts ...grpc.ServerOption) (*P4Device, error) {
	gnmiGrpcServer := grpc.NewServer(opts...)
	p4rtGrpcServer := grpc.NewServer(opts...)

	d := &P4Device{
		addr:           addr,
		gnmiGrpcServer: gnmiGrpcServer,
		gnmiServer:     gnmi.New(gnmiGrpcServer),
		p4rtGrpcServer: p4rtGrpcServer,
		p4rtServer:     p4rt.New(p4rtGrpcServer),
	}

	reflection.Register(gnmiGrpcServer)
	reflection.Register(p4rtGrpcServer)

	if err := d.startServers(); err != nil {
		return nil, fmt.Errorf("failed to start device: %v", err)
	}

	return d, nil
}

// Addr returns the currently configured address for the listening services.
func (d *P4Device) Addr() string {
	return d.addr
}

// Stop stops the listening services.
func (d *P4Device) Stop() {
	if d.stop == nil {
		return
	}
	d.stop()
}

// GNMI returns the gnmi server implementation.
func (d *P4Device) GNMI() *gnmi.Server {
	return d.gnmiServer
}

// P4RT returns the p4rt server implementation.
func (d *P4Device) P4RT() *p4rt.Server {
	return d.p4rtServer
}

func (d *P4Device) startServers() error {
	gnmiListener, err := net.Listen("tcp", d.addr+":"+GNMIPort)
	if err != nil {
		return fmt.Errorf("failed to start gnmi listener: %v", err)
	}
	p4rtListener, err := net.Listen("tcp", d.addr+":"+P4RTPort)
	if err != nil {
		return fmt.Errorf("failed to start p4runtime listener: %v", err)
	}

	// start services
	go d.gnmiGrpcServer.Serve(gnmiListener)
	go d.p4rtGrpcServer.Serve(p4rtListener)

	d.stop = func() {
		d.gnmiGrpcServer.Stop()
		gnmiListener.Close()
		d.p4rtGrpcServer.Stop()
		p4rtListener.Close()
	}

	return nil
}
