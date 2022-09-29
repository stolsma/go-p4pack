package dpdkicli

import (
	"github.com/spf13/cobra"
)

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
