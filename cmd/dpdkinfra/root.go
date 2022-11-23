// SPDX-FileCopyrightText: 2022-present Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"github.com/spf13/cobra"
)

// Create CLI handler
func CreateCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "",
		Short: "DPDKInfra is a Go/DPDK SWX example program",
		Long:  `Testing Go with DPDK - SWX. Complete documentation is available at https://github.com/stolsma/go-p4pack/`,
		Run:   func(cmd *cobra.Command, args []string) {},
	}
	var config, dpdkargs string
	root.Flags().StringVarP(&config, "config", "c", "./examples/default/config.json", "The config file to use.")
	// "dummy -c 3 -n 4"
	// "dummy -c 3 --log-level .*,8"
	root.Flags().StringVarP(&dpdkargs, "dpdkargs", "d", "dummy -c 3 -n 4", "The DPDK arguments to use.")

	return root
}
