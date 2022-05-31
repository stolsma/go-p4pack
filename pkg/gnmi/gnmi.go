// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package gnmi

import (
	"context"
	"time"

	gnmipb "github.com/openconfig/gnmi/proto/gnmi"
	"google.golang.org/grpc"
	grpcCodes "google.golang.org/grpc/codes"
	grpcStatus "google.golang.org/grpc/status"
)

// Server is the gNMI server implementation.
type Server struct {
	grpcServer   *grpc.Server
	subscription int
	Responses    [][]*gnmipb.SubscribeResponse
	GetResponses []interface{}
	Errs         []error
}

// New returns a new gNMI server instance.
func New(s *grpc.Server) *Server {
	srv := &Server{
		grpcServer: s,
	}

	gnmipb.RegisterGNMIServer(s, srv)
	return srv
}

func (s *Server) Capabilities(ctx context.Context, req *gnmipb.CapabilityRequest) (*gnmipb.CapabilityResponse, error) {
	return nil, grpcStatus.Errorf(grpcCodes.Unimplemented, "Unimplemented")
}

func (s *Server) Get(ctx context.Context, req *gnmipb.GetRequest) (*gnmipb.GetResponse, error) {
	if len(s.GetResponses) == 0 {
		return nil, grpcStatus.Errorf(grpcCodes.Unimplemented, "Unimplemented")
	}

	resp := s.GetResponses[0]
	s.GetResponses = s.GetResponses[1:]

	switch v := resp.(type) {
	case error:
		return nil, v
	case *gnmipb.GetResponse:
		return v, nil
	default:
		return nil, grpcStatus.Errorf(grpcCodes.DataLoss, "Unknown message type: %T", resp)
	}
}

func (s *Server) Set(ctx context.Context, req *gnmipb.SetRequest) (*gnmipb.SetResponse, error) {
	return nil, grpcStatus.Errorf(grpcCodes.Unimplemented, "Unimplemented")
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
