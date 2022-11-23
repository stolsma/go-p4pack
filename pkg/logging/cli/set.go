// SPDX-FileCopyrightText: 2020-2022 Open Networking Foundation <info@opennetworking.org>
// SPDX-FileCopyrightText: 2022-present Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"github.com/spf13/cobra"
	"github.com/stolsma/go-p4pack/pkg/logging"
)

func getSetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set",
		Short: "Sets a logger attribute (e.g. level)",
	}
	cmd.AddCommand(getSetLevelCommand())
	return cmd
}

func getSetLevelCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "level logger_name",
		Short: "Sets a logger level",
		Args:  cobra.ExactArgs(2),
		Run:   runSetLevelCommand,
	}

	return cmd
}

func runSetLevelCommand(cmd *cobra.Command, args []string) {
	name := args[0]
	if name == "" {
		cmd.PrintErrf("The logger name should be provided\n")
		return
	}

	levelArg := args[1]
	if levelArg == "" {
		cmd.PrintErrf("The logger level should be provided\n")
		return
	}

	// check if level really exists
	level := logging.LevelString2Level(levelArg)
	if level == logging.LastLevel {
		cmd.PrintErrf("The logger level should be one of: %s\n", logging.LevelStrings)
		return
	}

	// get the loggers operational configuration
	list := logging.GetLoggerDataList()
	if _, ok := list[name]; !ok {
		cmd.PrintErrf("The logger name does not exist!\n")
		return
	}

	logger := logging.GetLogger(name)
	logger.SetLevel(level)
	cmd.Printf("%s logger level is set to %s\n", name, level)
}
