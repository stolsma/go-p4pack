package dpdkicli

import (
	"github.com/spf13/cobra"
)

func initTap(parent *cobra.Command) {
	tap := &cobra.Command{
		Use:   "tap",
		Short: "Base command for all tap actions",
		// Run:   func(cmd *cobra.Command, args []string) {},
	}

	tap.AddCommand(&cobra.Command{
		Use:   "create [tapname]",
		Short: "Create a tap interface on the system",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			dpdki := getDpdki(cmd)
			err := dpdki.TapCreate(args[0])
			if err != nil {
				cmd.PrintErrf("TAP %s create err: %d\n", args[0], err)
			}
			cmd.Printf("TAP %s created!\n", args[0])
		},
	})

	show := &cobra.Command{
		Use:     "show [tapname]",
		Short:   "Show information of all (or one given) TAP interface(s)",
		Aliases: []string{"sh"},
		Args:    cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			dpdki := getDpdki(cmd)
			t := ""
			if len(args) == 1 {
				t = args[0]
			}
			list, err := dpdki.TapList(t)
			if err != nil {
				cmd.PrintErrf("TAP %s show err: %d\n", t, err)
			}
			cmd.Printf("Known TAP interfaces:\n%s", list)
		},
	}
	var re, li, si bool
	show.Flags().BoolVarP(&re, "repeat", "r", false, "Continuously update statistics (every second), use CTRL-C to stop.")
	show.Flags().BoolVarP(&li, "long", "l", false, "Show all information known for TAP interfaces.")
	show.Flags().BoolVarP(&si, "short", "s", true, "Show minimum information known for TAP interfaces.")
	show.MarkFlagsMutuallyExclusive("long", "short")
	tap.AddCommand(show)

	parent.AddCommand(tap)
}
