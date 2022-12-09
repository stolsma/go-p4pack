// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"sort"

	"github.com/spf13/cobra"
	"github.com/stolsma/go-p4pack/pkg/dpdkinfra"
)

func initInterface(parent *cobra.Command) {
	interf := &cobra.Command{
		Use:     "interface",
		Short:   "Base command for all interface actions",
		Aliases: []string{"int"},
	}

	initStats(interf)
	initLinkUpDown(interf)

	initTap(interf)
	initPmd(interf)
	parent.AddCommand(interf)
}

func initStats(parent *cobra.Command) {
	stats := &cobra.Command{
		Use:     "stats [name]",
		Short:   "Show statistics of all (or one given) interface(s)",
		Aliases: []string{"st"},
		Args:    cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			dpdki := dpdkinfra.Get()
			t := ""
			if len(args) == 1 {
				t = args[0]
			}

			stats, err := dpdki.GetPortStatsString(t)
			if err != nil {
				cmd.PrintErrf("Interface %v stats err: %v\n", t, err)
			}

			var names []string
			for name := range stats {
				names = append(names, name)
			}
			sort.Strings(names)

			for _, name := range names {
				cmd.Printf("%v", stats[name])
			}
		},
	}
	var re, li, si bool
	stats.Flags().BoolVarP(&re, "repeat", "r", false, "Continuously update statistics (every second), use CTRL-C to stop.")
	stats.Flags().BoolVarP(&li, "long", "l", false, "Show all information known for interfaces.")
	stats.Flags().BoolVarP(&si, "short", "s", true, "Show minimum information known for interfaces.")
	stats.MarkFlagsMutuallyExclusive("long", "short")

	parent.AddCommand(stats)
}

func initLinkUpDown(parent *cobra.Command) {
	lud := &cobra.Command{
		Use:     "link [name] [up/down]",
		Short:   "Set the interface up or down",
		Aliases: []string{"set"},
		Args:    cobra.MaximumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			dpdki := dpdkinfra.Get()
			var err error
			ud := ""
			t := ""
			if len(args) == 2 {
				t = args[0]
				ud = args[1]
			}

			switch ud {
			case "up":
				err = dpdki.LinkUp(t)
			case "down":
				err = dpdki.LinkDown(t)
			default:
				cmd.PrintErrf("Use up or down and not %v !\n", ud)
				return
			}

			if err != nil {
				cmd.PrintErrf("Interface %v set up/down err: %v\n", t, err)
				return
			}
			cmd.Printf("Interface link state changed to: %v\n", ud)
		},
	}
	parent.AddCommand(lud)
}

func initPmd(parent *cobra.Command) {
	pmd := &cobra.Command{
		Use:   "pmd",
		Short: "Base command for all pmd actions",
	}

	initPmdShow(pmd)
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
				cmd.PrintErrf("PMD %v show err: %v\n", t, err)
			}
			cmd.Printf("Known PMD interfaces:\n%v", list)
		},
	}
	var re, li, si bool
	show.Flags().BoolVarP(&re, "repeat", "r", false, "Continuously update statistics (every second), use CTRL-C to stop.")
	show.Flags().BoolVarP(&li, "long", "l", false, "Show all information known for PMD interfaces.")
	show.Flags().BoolVarP(&si, "short", "s", true, "Show minimum information known for PMD interfaces.")
	show.MarkFlagsMutuallyExclusive("long", "short")

	parent.AddCommand(show)
}
