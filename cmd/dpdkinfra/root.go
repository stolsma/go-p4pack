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
		Run: func(cmd *cobra.Command, args []string) {
		},
	}
	var spec, dpdkargs string
	root.Flags().StringVarP(&spec, "spec", "s", "./examples/default/default.spec", "The switch .spec file to use.")
	// "dummy -c 3 -n 4"
	// "dummy -c 3 --log-level .*,8"
	root.Flags().StringVarP(&dpdkargs, "dpdkargs", "d", "dummy -c 3 -n 4", "The DPDK arguments to use.")

	return root
}
