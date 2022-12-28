// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"sort"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/stolsma/go-p4pack/pkg/dpdkinfra"
	"github.com/stolsma/go-p4pack/pkg/dpdkswx/ethdev"
	"github.com/stolsma/go-p4pack/pkg/dpdkswx/pktmbuf"
	"github.com/stolsma/go-p4pack/pkg/dpdkswx/tap"
)

func interfaceCreateCmd(parent *cobra.Command) *cobra.Command {
	createCmd := &cobra.Command{
		Use:     "create",
		Short:   "Base command for all interface create actions",
		Aliases: []string{"cr"},
	}

	interfaceCreateTapCmd(createCmd)
	interfaceCreateEthdevCmd(createCmd)
	parent.AddCommand(createCmd)

	return createCmd
}

func interfaceCreateTapCmd(parent *cobra.Command) *cobra.Command {
	tapCmd := &cobra.Command{
		Use:   "tap [name] [pktmbuf] [mtu]",
		Short: "Create a tap interface on the system",
		Args:  cobra.ExactArgs(3),
		ValidArgsFunction: ValidateArguments(
			AppendHelp("You must choose a name for the tap interface you are adding"),
			completePktmbufArg,
			AppendHelp("You must specify the MTU for the tap interface you are adding"),
			AppendLastHelp(3, "This command does not take any more arguments"),
		),
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
	}

	parent.AddCommand(tapCmd)

	return tapCmd
}

func interfaceCreateEthdevCmd(parent *cobra.Command) *cobra.Command {
	ethdevCmd := &cobra.Command{
		Use:   "ethdev [name] [device] [pktmbuf] [# tx queues] [tx queuesize] [# rx queues] [rx queuesize] [mtu] [promiscuous]",
		Short: "Create an ethdev interface on the system",
		Args:  cobra.MatchAll(cobra.MinimumNArgs(7), cobra.MaximumNArgs(9)),
		ValidArgsFunction: ValidateArguments(
			AppendHelp("You must choose a name for the ethdev interface you are adding"),
			AppendHelp("You must specify the device name for the ethdev interface you are adding (i.e like 0000:04:00.1)"),
			completePktmbufArg,
			AppendHelp("You must specify the number of transmit queues for the ethdev interface you are adding"),
			AppendHelp("You must specify the transmit queuesize for the ethdev interface you are adding"),
			AppendHelp("You must specify the number of receive queues for the ethdev interface you are adding"),
			AppendHelp("You must specify the receive queuesize for the ethdev interface you are adding"),
			AppendHelp("You must specify the MTU for the ethdev interface you are adding (0/none is default value of interface)"),
			AppendHelp("You must specify the promiscuous mode for the tap interface you are adding (on/off, on is default)"),
			AppendLastHelp(9, "This command does not take any more arguments"),
		),
		Run: func(cmd *cobra.Command, args []string) {
			dpdki := dpdkinfra.Get()
			var params ethdev.Params

			// get device name
			params.DevName = args[1]

			// get pktmbuf
			params.Rx.Mempool = dpdki.PktmbufStore.Get(args[2])
			if params.Rx.Mempool == nil {
				cmd.PrintErrf("Pktmbuf %s not defined!\n", args[2])
				return
			}

			// get # TX Queues
			ntxq, err := strconv.ParseInt(args[3], 0, 16)
			if err != nil {
				cmd.PrintErrf("# TX Queues (%s) is not a correct integer: %d\n", args[3], err)
				return
			}
			params.Tx.NQueues = uint16(ntxq)

			// get TX Queuesize
			txqsize, err := strconv.ParseInt(args[4], 0, 32)
			if err != nil {
				cmd.PrintErrf("TX Queuesize (%s) is not a correct integer: %d\n", args[4], err)
				return
			}
			params.Tx.QueueSize = uint32(txqsize)

			// get # RX Queues
			nrxq, err := strconv.ParseInt(args[5], 0, 16)
			if err != nil {
				cmd.PrintErrf("# RX Queues (%s) is not a correct integer: %d\n", args[5], err)
				return
			}
			params.Rx.NQueues = uint16(nrxq)

			// get RX Queuesize
			rxqsize, err := strconv.ParseInt(args[6], 0, 32)
			if err != nil {
				cmd.PrintErrf("RX Queuesize (%s) is not a correct integer: %d\n", args[6], err)
				return
			}
			params.Tx.QueueSize = uint32(rxqsize)

			// get MTU if available
			var mtu int64 = 1500
			if len(args) > 7 {
				mtu, err = strconv.ParseInt(args[7], 0, 16)
				if err != nil {
					cmd.PrintErrf("MTU (%s) is not a correct integer: %d\n", args[7], err)
					return
				}
				if mtu == 0 {
					mtu = 1500
				}
			}
			params.Rx.Mtu = uint16(mtu)

			// get promiscuous mode if available, else default value
			var prom = true
			if len(args) > 8 {
				val := strings.ToLower(args[8])
				switch val {
				case "on":
					prom = true
				case "off":
					prom = false
				default:
					cmd.PrintErrf("Promiscuous mode value (%s) should be `on` or `off`\n", args[8])
					return
				}
			}
			params.Promiscuous = prom

			// TODO Rx RSS needs to be implemented!
			//	p.Rx.Rss = vh.Rx.Rss

			// create
			_, err = dpdki.EthdevCreate(args[0], &params)
			if err != nil {
				cmd.PrintErrf("Ethdev %s create err: %d\n", args[0], err)
				return
			}

			cmd.Printf("Ethdev %s created!\n", args[0])
		},
	}

	parent.AddCommand(ethdevCmd)

	return ethdevCmd
}

// complete a pktmbuf argument
func completePktmbufArg(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	var directive = cobra.ShellCompDirectiveNoFileComp // | cobra.ShellCompDirectiveNoSpace

	// get pktmbuf list
	listPktmbuf := pktmbufList()

	// filter list with string to complete
	completions := filterCompletions(listPktmbuf, toComplete, &directive, "No Pktmbufs available for completion!")

	return completions, directive
}

// retrieve all pktmbuf names and return in sorted list.
func pktmbufList() []string {
	dpdki := dpdkinfra.Get()

	list := []string{}
	dpdki.PktmbufStore.Iterate(func(key string, value *pktmbuf.Pktmbuf) error {
		list = append(list, key)
		return nil
	})

	sort.Strings(list)

	return list
}
