// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"github.com/spf13/cobra"
	"github.com/stolsma/go-p4pack/pkg/dpdkinfra"
)

func interfaceCmd(parent *cobra.Command) *cobra.Command {
	interfaceCmd := &cobra.Command{
		Use:     "interface",
		Short:   "Base command for all interface actions",
		Aliases: []string{"int"},
	}

	interfaceDeviceCmd(interfaceCmd)
	interfaceCreateCmd(interfaceCmd)
	interfaceShowCmd(interfaceCmd)
	interfaceStatsCmd(interfaceCmd)
	interfaceLinkUpDownCmd(interfaceCmd)

	parent.AddCommand(interfaceCmd)

	return interfaceCmd
}

func interfaceLinkUpDownCmd(parent *cobra.Command) *cobra.Command {
	ludCmd := &cobra.Command{
		Use:     "link [name] [up/down]",
		Short:   "Set the interface up or down",
		Aliases: []string{"set"},
		Args:    cobra.MaximumNArgs(2),
		ValidArgsFunction: ValidateArguments(
			completePortList,
			AppendHelp("Set the interface state to up or down (up/down)"),
			AppendLastHelp(2, "This command does not take any more arguments"),
		),
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
