// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"io"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	shell "github.com/stolsma/go-p4pack/pkg/sshshell"
)

type cliHandler struct {
	s   *shell.Shell
	cli *cobra.Command
}

func createHandlerFactory(createRoot func() *cobra.Command) shell.HandlerFactory {
	return func(s *shell.Shell) shell.Handler {
		rw := s.GetReadWrite()  // get the read/write stream of this shell session
		cliRoot := createRoot() // create a new cli root for each session
		cliRoot.SetOut(rw)      // set output stream
		cliRoot.SetErr(rw)      // set error stream
		cliRoot.SetIn(rw)       // set input stream
		return &cliHandler{
			s:   s,
			cli: cliRoot,
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
	log.Infof("LINE from %s: %s", h.s.InstanceName(), line)

	// clean annotations and flags
	h.cli.Annotations = make(map[string]string)
	resetFlags(h.cli)

	// execute given textline
	h.cli.SetArgs(strings.Split(line, " "))
	h.cli.ExecuteContext(ctx)

	// executed command was exit?
	if h.cli.Annotations["exit"] != "" && h.cli.Annotations["exit"] == "exit" {
		return io.EOF
	}

	return nil
}

// Handle EOF input from remote user/app
func (h *cliHandler) HandleEOF() error {
	log.Infof("EOF from %s", h.s.InstanceName())
	return nil
}

// Start a SSH server with given context and DPDK Infra API
func startSSHShell(ctx context.Context, createRoot func() *cobra.Command) {
	go func(ctx context.Context, createRoot func() *cobra.Command) {
		sshServer := &shell.SSHServer{
			Config: &shell.Config{
				Bind: ":2222",
				Users: map[string]shell.User{
					"user": {Password: "notsecure"},
				},
				HostKeyFile: "./hostkey",
			},
			HandlerFactory: createHandlerFactory(createRoot),
		}

		// block and listen for ssh sessions and handle them
		err := sshServer.Listen(ctx)
		if err != nil {
			log.Errorf("sshServer ListenAndServe err:", err)
		}
	}(ctx, createRoot)
}
