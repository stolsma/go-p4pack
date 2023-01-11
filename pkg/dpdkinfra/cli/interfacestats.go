// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"sort"

	"github.com/spf13/cobra"
	"github.com/stolsma/go-p4pack/pkg/dpdkinfra"
)

func interfaceStatsCmd(parent *cobra.Command) *cobra.Command {
	var re bool
	statsCmd := &cobra.Command{
		Use:     "stats [name]",
		Short:   "Show statistics of all (or one given) interface(s)",
		Aliases: []string{"st"},
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

			stats, err := dpdki.GetPortStats(t)
			if err != nil {
				cmd.PrintErrf("Interface %v stats err: %v\n", t, err)
			}

			var names []string
			for name := range stats {
				names = append(names, name)
			}
			sort.Strings(names)

			for _, name := range names {
				hd := stats[name]["header"]
				cmd.Printf("\nInterface: %v <%v rx: %v tx: %v>\n", name, hd["type"], hd["rxqueuebound"], hd["txqueuebound"])

				if _, ok := stats[name]["err"]; ok {
					cmd.Printf("    Statistics: %v\n", stats[name]["err"]["err"])
				} else {
					st := stats[name]["stats"]
					cmd.Print("    Statistics:\n")

					switch hd["type"] {
					case "PMD":
						cmd.Printf("\tRX packets: %s bytes : %s\n", st["ipackets"], st["ibytes"])
						cmd.Printf("\tRX errors : %s missed: %s RX no mbuf: %s\n", st["ierrors"], st["imissed"], st["rxnombuf"])
						cmd.Printf("\tTX packets: %s bytes : %s\n", st["opackets"], st["obytes"])
						cmd.Printf("\tTX errors : %s\n", st["oerrors"])
					case "TAP":
						cmd.Print("\tno stats\n")
					}
				}
			}
		},
	}
	statsCmd.Flags().BoolVarP(&re, "repeat", "r", false, "Continuously update statistics (every second), use CTRL-C to stop.")

	parent.AddCommand(statsCmd)

	return statsCmd
}
