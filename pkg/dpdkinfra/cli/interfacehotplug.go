// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package cli

import "github.com/spf13/cobra"

func interfaceHotplugCmd(parent *cobra.Command) *cobra.Command {
	hotplugCmd := &cobra.Command{
		Use:     "hotplug",
		Short:   "Base command for all ethdev hotplug actions",
		Aliases: []string{"hp"},
	}

	interfaceHotplugAddCmd(hotplugCmd)
	interfaceHotplugListCmd(hotplugCmd)
	parent.AddCommand(hotplugCmd)

	return hotplugCmd
}

func interfaceHotplugListCmd(parent *cobra.Command) *cobra.Command {
	listCmd := &cobra.Command{
		Use:     "list",
		Short:   "List all ethdev (hotplug) interfaces on the system",
		Aliases: []string{"l"},
		Run: func(cmd *cobra.Command, args []string) {
		},
	}

	parent.AddCommand(listCmd)

	return listCmd
}

func interfaceHotplugAddCmd(parent *cobra.Command) *cobra.Command {
	addCmd := &cobra.Command{
		Use:   "add [device name] [devargs]",
		Short: "Create a ethdev hotplug interface on the system",
		Args:  cobra.ExactArgs(3),
		ValidArgsFunction: ValidateArguments(
			AppendHelp("You must choose a name for the tap interface you are adding"),
			completePktmbufArg,
			AppendHelp("You must specify the MTU for the tap interface you are adding"),
			AppendLastHelp(3, "This command does not take any more arguments"),
		),
		Run: func(cmd *cobra.Command, args []string) {
		},
	}

	parent.AddCommand(addCmd)

	return addCmd
}
