// Copyright 2023 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"github.com/spf13/cobra"
	"github.com/stolsma/go-p4pack/pkg/cli"
	"github.com/stolsma/go-p4pack/pkg/pcidevices"
)

func PciCmd(parents ...*cobra.Command) *cobra.Command {
	var pciCmd = &cobra.Command{
		Use:   "pci",
		Short: "Base command for all PCI device actions",
	}

	PciListCmd(pciCmd)
	PciBindCmd(pciCmd)
	PciUnbindCmd(pciCmd)
	return cli.AddCommand(parents, pciCmd)
}

func PciListCmd(parents ...*cobra.Command) *cobra.Command {
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all available PCI devices",
		Run: func(cmd *cobra.Command, args []string) {
			devices, err := pcidevices.GetPciDevices(pcidevices.AllDevices)
			if err != nil {
				cmd.PrintErrf("List err: %v\n", err)
			}

			cmd.Printf("%-13.13s %-30.30s %-27.27s %-27.27s %-15s\n", "Slot", "Class", "Vendor", "Device", "Driver")
			for _, d := range devices {
				cmd.Printf("%-13.13s %-30.30s %-27.27s %-27.27s %-15.15s\n", d.ID(), d.ClassExt(), d.VendorExt(), d.DeviceExt(), d.Driver())
			}
		},
	}

	return cli.AddCommand(parents, listCmd)
}

func PciBindCmd(parents ...*cobra.Command) *cobra.Command {
	bindCmd := &cobra.Command{
		Use:   "bind [pci device id] [driver]",
		Short: "bind PCI device to given driver",
		Args:  cobra.ExactArgs(2),
		ValidArgsFunction: cli.ValidateArguments(
			cli.AppendHelp("You must choose a pci address for the pci device you want to bind"),
			cli.AppendHelp("You must specify the name of the driver you want to bind to"),
			cli.AppendLastHelp(2, "This command does not take any more arguments"),
		),
		Run: func(cmd *cobra.Command, args []string) {
			device, err := pcidevices.New(args[0])
			if err != nil {
				cmd.PrintErrf("Get device error: %v\n", err)
				return
			}

			if err := device.Bind(args[1]); err != nil {
				cmd.PrintErrf("Device bind error: %v\n", err)
				return
			}

			cmd.Printf("Device %s is now bound to driver %s\n", args[0], args[1])
		},
	}

	return cli.AddCommand(parents, bindCmd)
}

func PciUnbindCmd(parents ...*cobra.Command) *cobra.Command {
	bindCmd := &cobra.Command{
		Use:   "unbind [pci device id]",
		Short: "Unbind PCI device from current driver",
		Args:  cobra.ExactArgs(1),
		ValidArgsFunction: cli.ValidateArguments(
			cli.AppendHelp("You must choose a pci address for the pci device you want to unbind"),
			cli.AppendLastHelp(1, "This command does not take any more arguments"),
		),
		Run: func(cmd *cobra.Command, args []string) {
			device, err := pcidevices.New(args[0])
			if err != nil {
				cmd.PrintErrf("Get device error: %v\n", err)
				return
			}

			if err := device.Unbind(); err != nil {
				cmd.PrintErrf("Device unbind error: %v\n", err)
				return
			}

			cmd.Printf("Device %s is now not bound to any driver!\n", args[0])
		},
	}

	return cli.AddCommand(parents, bindCmd)
}
