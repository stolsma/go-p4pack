package cli

import (
	"github.com/spf13/cobra"
	"github.com/stolsma/go-p4pack/pkg/dpdkinfra"
)

func initInterface(parent *cobra.Command) {
	interf := &cobra.Command{
		Use:     "interface",
		Short:   "Base command for all interface actions",
		Aliases: []string{"int"},
		// Run:     func(cmd *cobra.Command, args []string) {},
	}

	initTap(interf)
	initPmd(interf)
	parent.AddCommand(interf)
}

func initPmd(parent *cobra.Command) {
	pmd := &cobra.Command{
		Use:   "pmd",
		Short: "Base command for all pmd actions",
		// Run:   func(cmd *cobra.Command, args []string) {},
	}

	initPmdShow(pmd)
	initLinkUpDown(pmd)
	parent.AddCommand(pmd)
}

func initPmdShow(parent *cobra.Command) {
	show := &cobra.Command{
		Use:     "show [name]",
		Short:   "Show information of all (or one given) PMD interface(s)",
		Aliases: []string{"sh"},
		Args:    cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			dpdki := dpdkinfra.Get()
			t := ""
			if len(args) == 1 {
				t = args[0]
			}

			list, err := dpdki.EthdevList(t)
			if err != nil {
				cmd.PrintErrf("PMD %s show err: %d\n", t, err)
			}
			cmd.Printf("Known PMD interfaces:\n%s", list)
		},
	}
	var re, li, si bool
	show.Flags().BoolVarP(&re, "repeat", "r", false, "Continuously update statistics (every second), use CTRL-C to stop.")
	show.Flags().BoolVarP(&li, "long", "l", false, "Show all information known for PMD interfaces.")
	show.Flags().BoolVarP(&si, "short", "s", true, "Show minimum information known for PMD interfaces.")
	show.MarkFlagsMutuallyExclusive("long", "short")

	parent.AddCommand(show)
}

func initLinkUpDown(parent *cobra.Command) {
	lud := &cobra.Command{
		Use:     "link [up/down] [name]",
		Short:   "Set the PMD interface up or down",
		Aliases: []string{"set"},
		Args:    cobra.MaximumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			dpdki := dpdkinfra.Get()
			var err error
			ud := ""
			t := ""
			if len(args) == 2 {
				ud = args[0]
				t = args[1]
			}

			switch ud {
			case "up":
				err = dpdki.LinkUpDown(t, true)
			case "down":
				err = dpdki.LinkUpDown(t, false)
			default:
				cmd.PrintErrf("Use up or down and not %s !\n", ud)
				return
			}

			if err != nil {
				cmd.PrintErrf("PMD %s set up/down err: %d\n", t, err)
				return
			}
			cmd.Printf("PMD interface changed to: %s\n", ud)
		},
	}
	parent.AddCommand(lud)
}
