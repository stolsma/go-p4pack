// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package dpdkicli

import (
	"github.com/spf13/cobra"
)

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
