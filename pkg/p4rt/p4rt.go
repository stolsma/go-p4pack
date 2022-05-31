// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package p4rt

import (
	"google.golang.org/grpc"

	p4rtpb "github.com/p4lang/p4runtime/go/p4/v1"
)

// Server is a the p4rt server implementation.
type Server struct {
	*p4rtpb.UnimplementedP4RuntimeServer
	grpcServer *grpc.Server
}

// New returns a new p4rt server instance.
func New(s *grpc.Server) *Server {
	srv := &Server{
		grpcServer: s,
	}

	p4rtpb.RegisterP4RuntimeServer(s, srv)
	return srv
}
