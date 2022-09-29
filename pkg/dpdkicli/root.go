package dpdkicli

import (
	"io"

	"github.com/spf13/cobra"
	"github.com/stolsma/go-p4dpdk-vswitch/pkg/dpdkinfra"
)

func getDpdki(cmd *cobra.Command) *dpdkinfra.DpdkInfra {
	return cmd.Context().Value("dpdki").(*dpdkinfra.DpdkInfra)
}

// Create CLI handler
func Create(rw io.ReadWriter) *cobra.Command {
	root := &cobra.Command{
		Use:   "",
		Short: "DPDKInfra is a Go/DPDK SWX test program",
		Long:  `Testing Go with DPDK - SWX. Complete documentation is available at https://github.com/stolsma/p4vswitch/`,
		Run: func(cmd *cobra.Command, args []string) {
		},
	}

	// no completion create command
	root.CompletionOptions.DisableDefaultCmd = true

	// set in and output streams
	root.SetOut(rw)
	root.SetErr(rw)
	root.SetIn(rw)

	// add all supported root commands
	initVersion(root)
	initExit(root)
	initPipeline(root)
	initMempool(root)
	initInterface(root)

	return root
}
