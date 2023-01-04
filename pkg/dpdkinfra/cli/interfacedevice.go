// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"github.com/spf13/cobra"
	"github.com/stolsma/go-p4pack/pkg/dpdkinfra"
	"github.com/stolsma/go-p4pack/pkg/dpdkswx/ethdev"
)

func interfaceDeviceCmd(parent *cobra.Command) *cobra.Command {
	deviceCmd := &cobra.Command{
		Use:     "device",
		Short:   "Base command for all device actions",
		Aliases: []string{"dev"},
	}

	interfaceDeviceListCmd(deviceCmd)
	interfaceDeviceAttachCmd(deviceCmd)
	interfaceDeviceDetachCmd(deviceCmd)
	parent.AddCommand(deviceCmd)

	return deviceCmd
}

func interfaceDeviceListCmd(parent *cobra.Command) *cobra.Command {
	var used, notused bool
	listCmd := &cobra.Command{
		Use:     "list",
		Short:   "List all attached (hotplug) devices on the system",
		Aliases: []string{"l"},
		ValidArgsFunction: ValidateArguments(
			AppendLastHelp(0, "This command does not take any more arguments"),
		),
		Run: func(cmd *cobra.Command, args []string) {
			var devices []string
			var hs string

			if used { //nolint:gocritic
				devices = deviceList(UsedDevices)
				hs = "Used"
			} else if notused {
				devices = deviceList(UnusedDevices)
				hs = "Not used"
			} else {
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
	parent.AddCommand(listCmd)

	return listCmd
}

func interfaceDeviceAttachCmd(parent *cobra.Command) *cobra.Command {
	attachCmd := &cobra.Command{
		Use:     "attach [device argument string]",
		Short:   "Attach a DPDK device on the system via hotplug procedure",
		Aliases: []string{"a"},
		Args:    cobra.ExactArgs(1),
		ValidArgsFunction: ValidateArguments(
			AppendHelp("You must specify the DPDK device argument string for the interface you are attaching"),
			AppendLastHelp(1, "This command does not take any more arguments"),
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

	parent.AddCommand(attachCmd)

	return attachCmd
}

func interfaceDeviceDetachCmd(parent *cobra.Command) *cobra.Command {
	detachCmd := &cobra.Command{
		Use:     "detach [device name]",
		Short:   "Detach a DPDK device from the system via hotplug procedure",
		Aliases: []string{"a"},
		Args:    cobra.ExactArgs(1),
		ValidArgsFunction: ValidateArguments(
			completeUnusedDeviceList,
			AppendLastHelp(1, "This command does not take any more arguments"),
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

	parent.AddCommand(detachCmd)

	return detachCmd
}

type DeviceFilter uint

const (
	NoFilter DeviceFilter = iota
	AllDevices
	UsedDevices
	UnusedDevices
)

func completeUnusedDeviceList(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	var directive = cobra.ShellCompDirectiveNoFileComp // | cobra.ShellCompDirectiveNoSpace

	// get device list
	listDevice := deviceList(UnusedDevices)

	// filter list with string to complete
	completions := filterCompletions(listDevice, toComplete, &directive, "No devices available for completion!")

	return completions, directive
}

func deviceList(filter DeviceFilter) []string {
	var ports []*ethdev.Ethdev
	var key = make(map[string]bool)
	var err error
	dpdki := dpdkinfra.Get()
	list := []string{}

	// get list of attached EAL devices
	switch filter {
	case AllDevices:
		ports, err = dpdki.GetAttachedPorts()
	case UsedDevices:
		ports, err = dpdki.GetUsedPorts()
	case UnusedDevices:
		ports, err = dpdki.GetUnusedPorts()
	default:
		return list
	}

	if err != nil {
		return list
	}

	// copy devicenames and filter duplicates
	for _, port := range ports {
		devName := port.DevName()
		if !key[devName] {
			key[devName] = true
			list = append(list, devName)
		}
	}

	return list
}
