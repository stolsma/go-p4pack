// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"
	"github.com/stolsma/go-p4pack/pkg/dpdkinfra"
	"github.com/stolsma/go-p4pack/pkg/dpdkswx/ethdev"
)

func interfaceCmd(parent *cobra.Command) *cobra.Command {
	interfaceCmd := &cobra.Command{
		Use:     "interface",
		Short:   "Base command for all interface actions",
		Aliases: []string{"int"},
	}

	interfaceHotplugCmd(interfaceCmd)
	interfaceCreateCmd(interfaceCmd)
	interfaceShowCmd(interfaceCmd)
	interfaceStatsCmd(interfaceCmd)
	interfaceLinkUpDownCmd(interfaceCmd)

	interfacePmdCmd(interfaceCmd)
	parent.AddCommand(interfaceCmd)

	return interfaceCmd
}

func interfaceStatsCmd(parent *cobra.Command) *cobra.Command {
	statsCmd := &cobra.Command{
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
	statsCmd.Flags().BoolVarP(&re, "repeat", "r", false, "Continuously update statistics (every second), use CTRL-C to stop.")
	statsCmd.Flags().BoolVarP(&li, "long", "l", false, "Show all information known for interfaces.")
	statsCmd.Flags().BoolVarP(&si, "short", "s", true, "Show minimum information known for interfaces.")
	statsCmd.MarkFlagsMutuallyExclusive("long", "short")

	parent.AddCommand(statsCmd)

	return statsCmd
}

func interfaceLinkUpDownCmd(parent *cobra.Command) *cobra.Command {
	ludCmd := &cobra.Command{
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
	parent.AddCommand(ludCmd)

	return ludCmd
}

func interfacePmdCmd(parent *cobra.Command) *cobra.Command {
	pmdCmd := &cobra.Command{
		Use:   "pmd",
		Short: "Base command for all pmd actions",
	}

	interfacePmdShowCmd(pmdCmd)
	parent.AddCommand(pmdCmd)

	return pmdCmd
}

func interfacePmdShowCmd(parent *cobra.Command) *cobra.Command {
	showCmd := &cobra.Command{
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

			list := ""
			if err := dpdki.EthdevStore.Iterate(func(key string, ethdev *ethdev.Ethdev) error {
				list += fmt.Sprintf("  %s \n", ethdev.DevName())
				devInfo, err := ethdev.GetPortInfoString()
				list += fmt.Sprintf("%s \n", devInfo)
				list += "\n"
				return err
			}); err != nil {
				cmd.PrintErrf("PMD %v show err: %v\n", t, err)
			}

			cmd.Printf("Known PMD interfaces:\n%v", list)
		},
	}
	var re, li, si bool
	showCmd.Flags().BoolVarP(&re, "repeat", "r", false, "Continuously update statistics (every second), use CTRL-C to stop.")
	showCmd.Flags().BoolVarP(&li, "long", "l", false, "Show all information known for PMD interfaces.")
	showCmd.Flags().BoolVarP(&si, "short", "s", true, "Show minimum information known for PMD interfaces.")
	showCmd.MarkFlagsMutuallyExclusive("long", "short")

	parent.AddCommand(showCmd)

	return showCmd
}
