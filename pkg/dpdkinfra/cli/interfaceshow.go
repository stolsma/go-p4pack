// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"sort"

	"github.com/spf13/cobra"
	"github.com/stolsma/go-p4pack/pkg/dpdkinfra"
)

func interfaceShowCmd(parent *cobra.Command) *cobra.Command {
	showCmd := &cobra.Command{
		Use:     "show [portname]",
		Short:   "Show information of all (or one given) interface(s)",
		Aliases: []string{"sh"},
		Args:    cobra.MaximumNArgs(1),
		ValidArgsFunction: ValidateArguments(
			completePortList,
			AppendLastHelp(1, "This command does not take any more arguments"),
		),
		Run: func(cmd *cobra.Command, args []string) {
			dpdki := dpdkinfra.Get()
			t := ""
			if len(args) == 1 {
				t = args[0]
			}

			info, err := dpdki.GetPortInfo(t)
			if err != nil {
				cmd.PrintErrf("Interface %v info err: %v\n", t, err)
			}

			var names []string
			for name := range info {
				names = append(names, name)
			}
			sort.Strings(names)

			for _, name := range names {
				hd := info[name]["header"]
				cmd.Printf("\nInterface: %v <%v %v:%v %v:%v>\n", name, hd["type"],
					hd["pipein"], hd["pipeinport"], hd["pipeout"], hd["pipeoutport"])

				if _, ok := info[name]["err"]; ok {
					cmd.Printf("    Information: %v\n", info[name]["err"]["err"])
				} else {
					st := info[name]["info"]
					cmd.Print("\n")

					switch hd["type"] {
					case "PMD":
						cmd.Printf("  Status           : %s\n", st["status"])
						cmd.Printf("  Autonegotiation  : %s\n", st["autoneg"])
						cmd.Printf("  Duplex           : %s\n", st["duplex"])
						cmd.Printf("  Link speed       : %s\n", st["speed"])
						cmd.Printf("  Promiscuous mode : %s\n", st["promiscuous"])
						cmd.Printf("  MAC address      : %s\n", st["macaddr"])
						cmd.Print("\n")
						cmd.Print("  Port specific items:\n")
						cmd.Printf("%s\n", st["portinfo"])
					case "TAP":
						cmd.Print("\tno stats\n")
					}
				}
			}
		},
	}

	parent.AddCommand(showCmd)

	return showCmd
}
