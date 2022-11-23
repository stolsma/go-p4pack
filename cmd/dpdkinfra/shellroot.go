// SPDX-FileCopyrightText: 2022-present Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"github.com/spf13/cobra"
	dpdkinfracli "github.com/stolsma/go-p4pack/pkg/dpdkicli"
	flowtestcli "github.com/stolsma/go-p4pack/pkg/flowtest/cli"
	loggingcli "github.com/stolsma/go-p4pack/pkg/logging/cli"
)

// create SSH shell CLI handling interface
func createShellRoot() *cobra.Command {
	cliRoot := &cobra.Command{
		Use:   "",
		Short: "DPDKInfra is a Go/DPDK SWX pipeline test program",
		Long:  `Testing Go with DPDK SWX pipeline. Complete documentation is available at https://github.com/stolsma/go-p4pack/`,
		Run:   func(cmd *cobra.Command, args []string) {},
	}
	cliRoot.CompletionOptions.DisableDefaultCmd = true // no completion create command
	initExit(cliRoot)
	initVersion(cliRoot)
	dpdkinfracli.GetCommand(cliRoot)
	loggingcli.GetCommand(cliRoot)
	flowtestcli.GetCommand(cliRoot)
	return cliRoot
}

func initExit(parent *cobra.Command) {
	var exitCmd = &cobra.Command{
		Use:   "exit",
		Short: "exit",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Root().Annotations["exit"] = "exit"
		},
	}

	parent.AddCommand(exitCmd)
}

func initVersion(parent *cobra.Command) {
	parent.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print the version number of DPDKInfra",
		Long:  `All software has versions. This is DPDKInfra's`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("DPDKInfra v0.01 -- HEAD")
		},
	})
}
