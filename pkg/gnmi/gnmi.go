// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package gnmi

import (
	"context"
	"time"

	gnmipb "github.com/openconfig/gnmi/proto/gnmi"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Server is the gNMI server implementation.

type Server struct {
	name         string
	grpcServer   *grpc.Server
	subscription int

	Responses    [][]*gnmipb.SubscribeResponse
	GetResponses []interface{}
	Errs         []error
}

// New returns a new gNMI server instance.
func New(s *grpc.Server, name string) (*Server, error) {
	srv := &Server{
		name:       name,
		grpcServer: s,
	}

	gnmipb.RegisterGNMIServer(s, srv)
	return srv, nil
}

func (s *Server) Capabilities(ctx context.Context, req *gnmipb.CapabilityRequest) (*gnmipb.CapabilityResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "Unimplemented")
}

func (s *Server) Get(ctx context.Context, req *gnmipb.GetRequest) (*gnmipb.GetResponse, error) {
	if len(s.GetResponses) == 0 {
		return nil, status.Errorf(codes.Unimplemented, "Unimplemented")
	}

	resp := s.GetResponses[0]
	s.GetResponses = s.GetResponses[1:]

	switch v := resp.(type) {
	case error:
		return nil, v
	case *gnmipb.GetResponse:
		return v, nil
	default:
		return nil, status.Errorf(codes.DataLoss, "Unknown message type: %T", resp)
	}
}

func (s *Server) Set(ctx context.Context, req *gnmipb.SetRequest) (*gnmipb.SetResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "Unimplemented")
}

func (s *Server) Subscribe(stream gnmipb.GNMI_SubscribeServer) error {
	_, err := stream.Recv()
	if err != nil {
		return err
	}

	srs := s.Responses[s.subscription]
	if len(s.Errs) != s.subscription+1 {
		s.Errs = append(s.Errs, nil)
	}

	srErr := s.Errs[s.subscription]
	s.subscription++

	for _, sr := range srs {
		stream.Send(sr)
	}

	time.Sleep(5 * time.Second)
	return srErr
}
