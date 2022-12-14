// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"context"
	"fmt"
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

func (h *cliHandler) HandleCompletion(ctx context.Context, line string) (string, error) {
	log.Infof("Completion LINE from %s: %s", h.s.InstanceName(), line)

	// clean annotations and flags
	h.cli.Annotations = make(map[string]string)
	resetFlags(h.cli)

	// redirect pipes
	bufOut := new(bytes.Buffer) // buffer stdout output
	bufErr := new(bytes.Buffer) // buffer stderr output
	h.cli.SetOut(bufOut)        // set output stream
	h.cli.SetErr(bufErr)        // set error stream

	// extend given line with complete request command and execute
	h.cli.SetArgs(strings.Split(cobra.ShellCompRequestCmd+" "+line, " "))
	h.cli.ExecuteContext(ctx)

	// process output - parce directive and split returned completion hints + help
	args := strings.Split(bufOut.String(), "\n")
	directive := args[len(args)-2]
	args = args[0 : len(args)-2]

	// process completion
	if len(args) > 1 {
		// more then one hint, so print them
		h.s.Output(createHintHelp(args))
	} else if directive == ":4" && len(args) == 1 {
		// only one hint so replace last argument
		gArgs := strings.Split(line, " ")
		dumArgs := strings.Split(args[0], "\t")
		gArgs[len(gArgs)-1] = dumArgs[0] + " "
		line = strings.Join(gArgs, " ")
	}

	// put cli readers/writers back
	rw := h.s.GetReadWrite() // get the read/write stream of this shell session
	h.cli.SetOut(rw)         // set output stream
	h.cli.SetErr(rw)         // set error stream

	// return completion result
	return line, nil
}

func createHintHelp(args []string) string {
	list := map[string]string{}
	l := 0
	for _, arg := range args {
		// split hint command and possible included hint command help text
		sArg := strings.Split(arg, "\t")
		if len(sArg) > 1 {
			list[sArg[0]] = sArg[1]
		} else {
			list[sArg[0]] = ""
		}
		// get the longest hint command string length for formatting
		if len(sArg[0]) > l {
			l = len(sArg[0])
		}
	}

	result := ""
	for key, h := range list {
		result += fmt.Sprintf("%-*s %s\n", l, key, h)
	}
	return result
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
