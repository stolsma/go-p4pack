// SPDX-FileCopyrightText: 2022-present Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"github.com/spf13/cobra"
	"github.com/stolsma/go-p4pack/pkg/cli"
	"github.com/stolsma/go-p4pack/pkg/flowtest"
)

func FlowtestStartCmd(parents ...*cobra.Command) *cobra.Command {
	startCmd := &cobra.Command{
		Use:   "start [flowtest name]",
		Short: "Starts a defined flowtest or all if name omitted",
		Args:  cobra.MaximumNArgs(1),
		Run:   runStartFTCommand,
	}

	return cli.AddCommand(parents, startCmd)
}

func runStartFTCommand(cmd *cobra.Command, args []string) {
	var name string

	// get flowtest singleton
	ft := flowtest.Get()
	if ft == nil {
		cmd.PrintErrf("The flowtest module is not initialized yet!\n")
		return
	}

	// get name argument
	if len(args) == 1 {
		name = args[0]
	}

	// execute
	if name == "" {
		err := ft.StartAll()
		if err != nil {
			cmd.PrintErrf("Something went wrong starting all the flowtests: %d \n", err)
			return
		}
		cmd.Println("All defined flowtests are started!")
	} else {
		cmd.PrintErrf("The start [flowtest name] command is not implemented yet!\n")
	}
}
