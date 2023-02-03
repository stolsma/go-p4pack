// SPDX-FileCopyrightText: 2020-2022 Open Networking Foundation <info@opennetworking.org>
// SPDX-FileCopyrightText: 2022-present Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"github.com/spf13/cobra"
	"github.com/stolsma/go-p4pack/pkg/cli"
	"github.com/stolsma/go-p4pack/pkg/logging"
)

func LogGetCommand(parents ...*cobra.Command) *cobra.Command {
	getCmd := &cobra.Command{
		Use:   "get",
		Short: "Gets a logger attribute (e.g. level)",
	}

	LogGetLevelCommand(getCmd)
	return cli.AddCommand(parents, getCmd)
}

func LogGetLevelCommand(parents ...*cobra.Command) *cobra.Command {
	levelCmd := &cobra.Command{
		Use:   "level logger_name",
		Short: "Gets a logger level",
		Args:  cobra.ExactArgs(1),
		Run:   runGetLevelCommand,
	}

	return cli.AddCommand(parents, levelCmd)
}

func runGetLevelCommand(cmd *cobra.Command, args []string) {
	name := args[0]
	if name == "" {
		cmd.PrintErrf("The logger name should be provided\n")
		return
	}

	// get the loggers operational configuration
	list := logging.GetLoggerDataList()
	if _, ok := list[name]; !ok {
		cmd.PrintErrf("The logger name does not exist!\n")
		return
	}

	cmd.Printf("%s logger level is %s\n", name, list[name].Level)
}
