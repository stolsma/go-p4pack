// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stolsma/go-p4pack/pkg/dpdkinfra"
	"github.com/stolsma/go-p4pack/pkg/dpdkswx/tap"
)

func interfaceShowCmd(parent *cobra.Command) *cobra.Command {
	showCmd := &cobra.Command{
		Use:     "show",
		Short:   "Base command for all interface show actions",
		Aliases: []string{"cr"},
	}

	interfaceShowTapCmd(showCmd)
	parent.AddCommand(showCmd)

	return showCmd
}

func interfaceShowTapCmd(parent *cobra.Command) *cobra.Command {
	tapCmd := &cobra.Command{
		Use:     "tap [tapname]",
		Short:   "Show information of all (or one given) TAP interface(s)",
		Aliases: []string{"sh"},
		Args:    cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			dpdki := dpdkinfra.Get()
			t := ""
			if len(args) == 1 {
				t = args[0]
			}

			list := ""
			if err := dpdki.TapStore.Iterate(func(key string, tap *tap.Tap) error {
				list += fmt.Sprintf("  %s \n", tap.Name())
				return nil
			}); err != nil {
				cmd.PrintErrf("TAP %s show err: %d\n", t, err)
				return
			}

			cmd.Printf("Known TAP interfaces:\n%s", list)
		},
	}
	var re, li, si bool
	tapCmd.Flags().BoolVarP(&re, "repeat", "r", false, "Continuously update statistics (every second), use CTRL-C to stop.")
	tapCmd.Flags().BoolVarP(&li, "long", "l", false, "Show all information known for TAP interfaces.")
	tapCmd.Flags().BoolVarP(&si, "short", "s", true, "Show minimum information known for TAP interfaces.")
	tapCmd.MarkFlagsMutuallyExclusive("long", "short")

	parent.AddCommand(tapCmd)

	return tapCmd
}
