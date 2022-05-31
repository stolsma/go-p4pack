// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package p4device

import (
	"context"
	"testing"

	"google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	status "google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	// gNMI
	gnmipb "github.com/openconfig/gnmi/proto/gnmi"

	// P4rt
	p4rtpb "github.com/p4lang/p4runtime/go/p4/v1"
)

func TestGNMI(t *testing.T) {
	device, err := New("")
	if err != nil {
		t.Fatalf("failed to start p4device: %v", err)
	}
	defer device.Stop()

	conn, err := grpc.Dial(device.Addr()+":"+GNMIPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("failed to Dial p4device: %v", err)
	}

	want := &gnmipb.GetResponse{
		Notification: []*gnmipb.Notification{{
			Update: []*gnmipb.Update{{
				Path: &gnmipb.Path{
					Elem: []*gnmipb.PathElem{
						{Name: "intefaces"},
						{Name: "inteface", Key: map[string]string{"name": "eth0"}},
						{Name: "mtu"},
					},
				},
				Val: &gnmipb.TypedValue{
					Value: &gnmipb.TypedValue_IntVal{
						IntVal: 1500,
					},
				},
			}},
		}},
	}
	device.gnmiServer.GetResponses = []interface{}{want}

	cGNMI := gnmipb.NewGNMIClient(conn)
	resp, err := cGNMI.Get(context.Background(), &gnmipb.GetRequest{})
	if err != nil {
		t.Fatalf("gnmi.Get failed: %v", err)
	}

	if !proto.Equal(resp, want) {
		t.Fatalf("gnmi.Get failed got %v, want %v", resp, want)
	}
}

func TestP4RT(t *testing.T) {
	device, err := New("")
	if err != nil {
		t.Fatalf("failed to start p4device: %v", err)
	}
	defer device.Stop()

	conn, err := grpc.Dial(device.Addr()+":"+P4RTPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("failed to Dial p4device: %v", err)
	}

	want := &p4rtpb.CapabilitiesRequest{}

	cP4RT := p4rtpb.NewP4RuntimeClient(conn)
	_, err = cP4RT.Capabilities(context.Background(), want)
	if err == nil {
		t.Fatalf("p4rt.Capabilities worked and that shouldnt! ;-)")
	}

	if err != nil {
		e, _ := status.FromError(err)
		if e.Code() != codes.Unimplemented {
			t.Fatalf("Unexpected p4rt.Capabilities error: %v", err)
		}
	}
}
