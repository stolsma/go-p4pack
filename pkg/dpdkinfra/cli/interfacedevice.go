// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"github.com/spf13/cobra"
	"github.com/stolsma/go-p4pack/pkg/cli"
	"github.com/stolsma/go-p4pack/pkg/dpdkinfra"
)

func InterfaceDeviceCmd(parents ...*cobra.Command) *cobra.Command {
	deviceCmd := &cobra.Command{
		Use:     "device",
		Short:   "Base command for all device actions",
		Aliases: []string{"dev"},
	}

	InterfaceDeviceListCmd(deviceCmd)
	InterfaceDeviceAttachCmd(deviceCmd)
	InterfaceDeviceDetachCmd(deviceCmd)
	return cli.AddCommand(parents, deviceCmd)
}

func InterfaceDeviceListCmd(parents ...*cobra.Command) *cobra.Command {
	var used, notused bool
	listCmd := &cobra.Command{
		Use:     "list",
		Short:   "List all attached (hotplug) devices on the system",
		Aliases: []string{"l"},
		ValidArgsFunction: cli.ValidateArguments(
			cli.AppendLastHelp(0, "This command does not take any more arguments"),
		),
		Run: func(cmd *cobra.Command, args []string) {
			var devices []string
			var hs string

			switch {
			case used:
				devices = deviceList(UsedDevices)
				hs = "Used"
			case notused:
				devices = deviceList(UnusedDevices)
				hs = "Not used"
			default:
				devices = deviceList(AllDevices)
				hs = "All"
			}

			cmd.Printf("%s attached DPDK devices:\n", hs)
			for _, device := range devices {
				cmd.Printf("  %s\n", device)
			}
		},
	}
	listCmd.Flags().BoolVarP(&used, "used", "u", false, "Show all used DPDK devices.")
	listCmd.Flags().BoolVarP(&notused, "notused", "n", false, "Show all not used DPDK devices.")
	listCmd.MarkFlagsMutuallyExclusive("used", "notused")
	return cli.AddCommand(parents, listCmd)
}

func InterfaceDeviceAttachCmd(parents ...*cobra.Command) *cobra.Command {
	attachCmd := &cobra.Command{
		Use:     "attach [device argument string]",
		Short:   "Attach a DPDK device on the system via hotplug procedure",
		Aliases: []string{"a"},
		Args:    cobra.ExactArgs(1),
		ValidArgsFunction: cli.ValidateArguments(
			cli.AppendHelp("You must specify the DPDK device argument string for the interface you are attaching"),
			cli.AppendLastHelp(1, "This command does not take any more arguments"),
		),
		Run: func(cmd *cobra.Command, args []string) {
			dpdki := dpdkinfra.Get()

			devArgs, err := dpdki.AttachDevice(args[0])
			if err != nil {
				cmd.PrintErrf("Error creating device: %v", err)
				return
			}

			cmd.Printf("Requested device (%s) created via hotplug on bus %s!\n", devArgs.Name(), devArgs.Bus())
		},
	}

	return cli.AddCommand(parents, attachCmd)
}

func InterfaceDeviceDetachCmd(parents ...*cobra.Command) *cobra.Command {
	detachCmd := &cobra.Command{
		Use:     "detach [device name]",
		Short:   "Detach a DPDK device from the system via hotplug procedure",
		Aliases: []string{"a"},
		Args:    cobra.ExactArgs(1),
		ValidArgsFunction: cli.ValidateArguments(
			completeUnusedDeviceList,
			cli.AppendLastHelp(1, "This command does not take any more arguments"),
		),
		Run: func(cmd *cobra.Command, args []string) {
			dpdki := dpdkinfra.Get()

			_, err := dpdki.DetachDevice(args[0])
			if err != nil {
				cmd.PrintErrf("Error detaching device: %v\n", err)
				return
			}

			cmd.Printf("Device %s is detached!\n", args[0])
		},
	}

	return cli.AddCommand(parents, detachCmd)
}
