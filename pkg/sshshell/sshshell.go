// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package sshshell

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type cliHandler struct {
	s   *Shell
	cli *cobra.Command
}

func createHandlerFactory(createRoot func() *cobra.Command) HandlerFactory {
	return func(s *Shell) Handler {
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

const activeHelpMarker = "_activeHelp_ "

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
	hints := strings.Split(bufOut.String(), "\n")
	directive, err := parseDirective(hints[len(hints)-2])
	hints = hints[0 : len(hints)-2]

	// check if error happened
	if directive&cobra.ShellCompDirectiveError == 0 && err == nil {
		// no error so process completion
		if len(hints) > 1 {
			// more then one hint, so print them
			h.s.Output(createHintHelp(hints))
		} else if directive&cobra.ShellCompDirectiveNoFileComp != 0 {
			if len(hints) == 1 {
				// only one hint
				lineArgs := strings.Split(line, " ")
				if strings.Contains(hints[0], activeHelpMarker) {
					// its an active help response
					help := strings.Replace(hints[0], activeHelpMarker, "", 1)
					h.s.OutputLine(help)
				} else {
					// normal hint so replace last (partial) argument
					hint := strings.Split(hints[0], "\t")
					lineArgs[len(lineArgs)-1] = hint[0]
					line = strings.Join(lineArgs, " ")
				}
			}

			// don't add space at the end when requested
			if directive&cobra.ShellCompDirectiveNoSpace == 0 {
				line += " "
			}
		}
	}

	// put cli readers/writers back
	rw := h.s.GetReadWrite() // get the read/write stream of this shell session
	h.cli.SetOut(rw)         // set output stream
	h.cli.SetErr(rw)         // set error stream

	// return completion result
	return line, nil
}

func parseDirective(arg string) (cobra.ShellCompDirective, error) {
	directive, err := strconv.ParseInt(strings.Replace(arg, ":", "", 1), 10, 0)
	return cobra.ShellCompDirective(directive), err
}

func createHintHelp(args []string) string {
	list := map[string]string{}
	hints := []string{}
	l := 0
	for _, arg := range args {
		// split hint command and possible included hint command help text
		sArg := strings.Split(arg, "\t")
		if len(sArg) > 1 {
			list[sArg[0]] = sArg[1]
		} else {
			list[sArg[0]] = ""
		}

		// store hints for sorting
		hints = append(hints, sArg[0])

		// get the longest hint command string length for formatting
		if len(sArg[0]) > l {
			l = len(sArg[0])
		}
	}

	// sort hints
	sort.Strings(hints)

	result := ""
	for _, hint := range hints {
		result += fmt.Sprintf("  %-*s\t%s\n", l, hint, list[hint])
	}
	return result
}

// Handle EOF input from remote user/app
func (h *cliHandler) HandleEOF() error {
	log.Infof("EOF from %s", h.s.InstanceName())
	return nil
}

// Start a SSH server with given context and DPDK Infra API
func StartSSHShell(ctx context.Context, createRoot func() *cobra.Command, config *Config) {
	go func(ctx context.Context, createRoot func() *cobra.Command) {
		sshServer := &SSHServer{
			Config:         config,
			HandlerFactory: createHandlerFactory(createRoot),
		}

		// block and listen for ssh sessions and handle them
		err := sshServer.Listen(ctx)
		if err != nil {
			log.Errorf("sshServer ListenAndServe err:", err)
		}
	}(ctx, createRoot)
}
