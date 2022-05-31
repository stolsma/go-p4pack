// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	log "github.com/golang/glog"

	p4rtpb "github.com/p4lang/p4runtime/go/p4/v1"
	"google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	status "google.golang.org/grpc/status"

	"github.com/stolsma/go-p4dpdk-vswitch/pkg/p4device"
	p4DpdkTarget "github.com/stolsma/go-p4dpdk-vswitch/pkg/tdi/targets/p4-dpdk-target"
)

func waitForSignal() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	done := make(chan bool, 1)

	go func() {
		<-sigs
		done <- true
	}()

	<-done
}

// First start the device threads
// then add the supported targets to the device
func main() {
	fmt.Println("Hello, Modules!")
	fmt.Printf("Process pid: %d\n", os.Getpid())

	device, err := p4device.New("")
	if err != nil {
		log.Errorf("failed to start p4device: %v", err)
		return
	}
	defer device.Stop()

	conn, err := grpc.Dial(device.Addr()+":"+p4device.P4RTPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Errorf("failed to Dial p4device: %v", err)
		return
	}

	want := &p4rtpb.CapabilitiesRequest{}

	cP4RT := p4rtpb.NewP4RuntimeClient(conn)
	_, err = cP4RT.Capabilities(context.Background(), want)
	if err == nil {
		log.Errorf("p4rt.Capabilities worked and that shouldnt! ;-)")
		return
	}
	if err != nil {
		e, _ := status.FromError(err)
		if e.Code() != codes.Unimplemented {
			log.Errorf("Unexpected p4rt.Capabilities error: %v", err)
			return
		}
	}

	fmt.Println("p4runtime grpc call worked!")

	// Initialize the first supported target: p4-DPDK
	err = p4DpdkTarget.InitTarget()
	if err != nil {
		log.Errorf("Unexpected p4DpdkTarget.InitTarget error: %v", err)
		return
	}

	// wait for signals to react on
	waitForSignal()

	fmt.Println("Telnet stopped!")
}
