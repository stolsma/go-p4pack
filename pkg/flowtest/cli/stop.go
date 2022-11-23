// SPDX-FileCopyrightText: 2020-2022 Open Networking Foundation <info@opennetworking.org>
// SPDX-FileCopyrightText: 2022-present Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"github.com/spf13/cobra"
)

func getStopCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stop flowtest_name",
		Short: "Stops a running flowtest",
		Args:  cobra.ExactArgs(1),
		Run:   runStopFTCommand,
	}

	return cmd
}

func runStopFTCommand(cmd *cobra.Command, args []string) {
	name := args[0]
	if name == "" {
		cmd.PrintErrf("The flowtest name should be provided\n")
		return
	}

	// cmd.Printf("%s logger level is %s\n", name, level)
}
