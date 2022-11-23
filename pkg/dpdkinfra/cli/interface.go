package cli

import (
	"github.com/spf13/cobra"
)

func initInterface(parent *cobra.Command) {
	interf := &cobra.Command{
		Use:     "interface",
		Short:   "Base command for all interface actions",
		Aliases: []string{"int"},
		// Run:     func(cmd *cobra.Command, args []string) {},
	}

	initTap(interf)
	parent.AddCommand(interf)
}
