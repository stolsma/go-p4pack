// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/stolsma/go-p4pack/pkg/dpdkinfra"
	"github.com/stolsma/go-p4pack/pkg/dpdkswx/tap"
)

func initTap(parent *cobra.Command) {
	tapC := &cobra.Command{
		Use:   "tap",
		Short: "Base command for all tap actions",
	}

	tapC.AddCommand(&cobra.Command{
		Use:   "create [tapname] [pktmbuf] [mtu]",
		Short: "Create a tap interface on the system",
		Args:  cobra.ExactArgs(3),
		Run: func(cmd *cobra.Command, args []string) {
			dpdki := dpdkinfra.Get()
			var params tap.Params

			// get pktmbuf
			arg1 := args[1]
			params.Pktmbuf = dpdki.PktmbufStore.Get(arg1)
			if params.Pktmbuf == nil {
				cmd.PrintErrf("Pktmbuf %s not defined!\n", arg1)
				return
			}

			// get MTU
			mtu, err := strconv.ParseInt(args[2], 0, 16)
			if err != nil {
				cmd.PrintErrf("Mtu (%s) is not a correct integer: %d\n", args[2], err)
				return
			}
			params.Mtu = int(mtu)

			// create
			_, err = dpdki.TapCreate(args[0], &params)
			if err != nil {
				cmd.PrintErrf("TAP %s create err: %d\n", args[0], err)
				return
			}

			cmd.Printf("TAP %s created!\n", args[0])
		},
	})

	showC := &cobra.Command{
		Use:     "show [tapname]",
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
	showC.Flags().BoolVarP(&re, "repeat", "r", false, "Continuously update statistics (every second), use CTRL-C to stop.")
	showC.Flags().BoolVarP(&li, "long", "l", false, "Show all information known for TAP interfaces.")
	showC.Flags().BoolVarP(&si, "short", "s", true, "Show minimum information known for TAP interfaces.")
	showC.MarkFlagsMutuallyExclusive("long", "short")
	tapC.AddCommand(showC)

	parent.AddCommand(tapC)
}
