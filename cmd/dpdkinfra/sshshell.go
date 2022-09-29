// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"io"
	"log"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/stolsma/go-p4dpdk-vswitch/pkg/dpdkicli"
	"github.com/stolsma/go-p4dpdk-vswitch/pkg/dpdkinfra"
	shell "github.com/stolsma/go-p4dpdk-vswitch/pkg/sshshell"
)

type cliHandler struct {
	s     *shell.Shell
	cli   *cobra.Command
	dpdki *dpdkinfra.DpdkInfra
}

func createHandlerFactory(dpdki *dpdkinfra.DpdkInfra) shell.HandlerFactory {
	return func(s *shell.Shell) shell.Handler {
		return &cliHandler{
			s:     s,
			cli:   dpdkicli.Create(s.GetReadWrite()),
			dpdki: dpdki,
		}
	}
}

func resetFlags(cmd *cobra.Command) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		f.Changed = false
		f.Value.Set(f.DefValue)
	})
	for _, sub := range cmd.Commands() {
		resetFlags(sub)
	}
}

func (h *cliHandler) HandleLine(ctx context.Context, line string) error {
	log.Printf("LINE from %s: %s", h.s.InstanceName(), line)

	// set context with dpdki structure and create annotations map
	// TODO solve linting error in correct way
	//lint:ignore SA1029 needs to be solved later
	vCtx := context.WithValue(ctx, "dpdki", h.dpdki)
	h.cli.Annotations = make(map[string]string)

	// execute given textline
	resetFlags(h.cli)
	h.cli.SetArgs(strings.Split(line, " "))
	h.cli.ExecuteContext(vCtx)

	// executed command was exit?
	if h.cli.Annotations["exit"] != "" && h.cli.Annotations["exit"] == "exit" {
		return io.EOF
	} else {
		return nil
	}
}

// Handle EOF input from remote user/app
func (h *cliHandler) HandleEof() error {
	log.Printf("EOF from %s", h.s.InstanceName())
	return nil
}

// Start a SSH server with given context and DPDK Infra API
func startSsh(ctx context.Context, dpdki *dpdkinfra.DpdkInfra) {
	go func(ctx context.Context, dpdki *dpdkinfra.DpdkInfra) {
		sshServer := &shell.SshServer{
			Config: &shell.Config{
				Bind: ":2222",
				Users: map[string]shell.User{
					"user": {Password: "notsecure"},
				},
				HostKeyFile: "./hostkey",
			},
			HandlerFactory: createHandlerFactory(dpdki),
		}

		// block and listen for ssh sessions and handle them
		err := sshServer.Listen(ctx)
		if err != nil {
			log.Println("sshServer ListenAndServe err:", err)
		}
	}(ctx, dpdki)
}
