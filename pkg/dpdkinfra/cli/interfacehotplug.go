// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"github.com/spf13/cobra"
	"github.com/stolsma/go-p4pack/pkg/dpdkinfra"
)

func interfaceHotplugCmd(parent *cobra.Command) *cobra.Command {
	hotplugCmd := &cobra.Command{
		Use:     "hotplug",
		Short:   "Base command for all device actions",
		Aliases: []string{"hp"},
	}

	interfaceHotplugListCmd(hotplugCmd)
	interfaceHotplugAttachCmd(hotplugCmd)
	interfaceHotplugDetachCmd(hotplugCmd)
	parent.AddCommand(hotplugCmd)

	return hotplugCmd
}

func interfaceHotplugListCmd(parent *cobra.Command) *cobra.Command {
	listCmd := &cobra.Command{
		Use:     "list",
		Short:   "List all attached (hotplug) devices on the system",
		Aliases: []string{"l"},
		Run: func(cmd *cobra.Command, args []string) {
			dpdki := dpdkinfra.Get()

			// get list of attached EAL ports
			ports, err := dpdki.GetAttachedPorts()
			if err != nil {
				cmd.PrintErrf("Error reading attached port list: %v", err)
				return
			}

			// print portlist
			cmd.Println("Valid EAL Ports:")
			for _, port := range ports {
				cmd.Printf("  %s\n", port.Name())
			}
		},
	}

	parent.AddCommand(listCmd)

	return listCmd
}

func interfaceHotplugAttachCmd(parent *cobra.Command) *cobra.Command {
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

			devArgs, err := dpdki.HotplugAdd(args[0])
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

func interfaceHotplugDetachCmd(parent *cobra.Command) *cobra.Command {
	detachCmd := &cobra.Command{
		Use:     "detach [device name]",
		Short:   "Detach a DPDK device from the system via hotplug procedure",
		Aliases: []string{"a"},
		Args:    cobra.ExactArgs(1),
		ValidArgsFunction: ValidateArguments(
			completeDeviceList,
			AppendLastHelp(1, "This command does not take any more arguments"),
		),
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("Command not implemented yet!")
			// dpdki := dpdkinfra.Get()
			// cmd.Printf("Device %s detached!\n", args[0])
		},
	}

	parent.AddCommand(detachCmd)

	return detachCmd
}

func completeDeviceList(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	var directive = cobra.ShellCompDirectiveNoFileComp // | cobra.ShellCompDirectiveNoSpace

	// get device list
	listDevice := deviceList()

	// filter list with string to complete
	completions := filterCompletions(listDevice, toComplete, &directive, "No devices available for completion!")

	return completions, directive
}

func deviceList() []string {
	dpdki := dpdkinfra.Get()
	list := []string{}

	// get list of attached EAL devices
	ports, err := dpdki.GetAttachedPorts()
	if err != nil {
		return list
	}

	for _, port := range ports {
		list = append(list, port.DevName())
	}

	return list
}
