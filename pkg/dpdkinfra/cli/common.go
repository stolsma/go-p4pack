// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"io"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/stolsma/go-p4pack/pkg/dpdkinfra"
	"github.com/stolsma/go-p4pack/pkg/dpdkinfra/portmngr"
	"github.com/stolsma/go-p4pack/pkg/dpdkswx/ethdev"
	"github.com/stolsma/go-p4pack/pkg/dpdkswx/pktmbuf"
)

const (
	ETX = 0x3 // control-C
)

// wait for CTRL-C
func waitForCtrlC(input io.Reader) {
	buf := make([]byte, 1)
	for {
		amount, err := input.Read(buf)
		if err != nil {
			break
		}

		if amount > 0 {
			ch := buf[0]
			if ch == ETX {
				break
			}
		}
	}
}

func indent(chars string, orig string) string {
	var last bool
	interm := strings.Split(orig, "\n")
	if interm[len(interm)-1] == "" {
		interm = interm[:len(interm)-1]
		last = true
	}
	res := chars + strings.Join(interm, "\n"+chars)
	if last {
		return res + "\n"
	}
	return res
}

type ValidateFn = func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective)

// execute completion commands per argument
func ValidateArguments(fns ...ValidateFn) ValidateFn {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		var completions []string
		var directive cobra.ShellCompDirective
		var fnLen = len(fns)
		var argLen = len(args)

		// call completion function related to the current argument position
		if argLen <= fnLen-2 {
			completions, directive = fns[argLen](cmd, args, toComplete)
		}

		// if last real argument has no completion or if pos of current argument > max arguments call catch-all function
		if argLen >= fnLen-2 && len(completions) == 0 {
			comp, direct := fns[fnLen-1](cmd, args, toComplete)
			if len(comp) > 0 {
				completions = comp
				directive = direct
			}
		}

		return completions, directive
	}
}

// add active help text for argument, and show if argument is empty
func AppendHelp(helpTxt string) ValidateFn {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		var completions []string
		directive := cobra.ShellCompDirectiveNoFileComp

		// help needed
		if toComplete == "" {
			completions = cobra.AppendActiveHelp(completions, helpTxt)
			directive |= cobra.ShellCompDirectiveNoSpace
		}

		return completions, directive
	}
}

// add active help text shown when last valid argument is already entered
func AppendLastHelp(total int, helpTxt string) ValidateFn {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		var completions []string
		directive := cobra.ShellCompDirectiveNoFileComp

		// help needed
		if len(args) == total-1 && toComplete != "" || len(args) > total-1 {
			completions = cobra.AppendActiveHelp(completions, helpTxt)
			directive |= cobra.ShellCompDirectiveNoSpace
		}

		return completions, directive
	}
}

// filter all strings starting with string to complete
func filterCompletions(list []string, toComplete string, directive *cobra.ShellCompDirective, help string) []string {
	var completions []string

	// always no space
	*directive |= cobra.ShellCompDirectiveNoSpace

	// filter list
	for _, name := range list {
		if strings.HasPrefix(name, toComplete) {
			completions = append(completions, name)
		}
	}

	// if none filtered, select all strings
	if len(completions) == 0 {
		completions = list
	}

	// if only one answer left and that is same as toComplete return nothing
	if len(completions) == 1 && completions[0] == toComplete {
		completions = []string{}
		*directive ^= cobra.ShellCompDirectiveNoSpace
	} else if len(completions) == 0 {
		completions = cobra.AppendActiveHelp(completions, help)
	}

	return completions
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

type DeviceFilter uint

const (
	NoFilter DeviceFilter = iota
	AllDevices
	UsedDevices
	UnusedDevices
)

// handle completion of sorted list of unused device names
func completeUnusedDeviceList(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	var directive = cobra.ShellCompDirectiveNoFileComp // | cobra.ShellCompDirectiveNoSpace

	// get device list
	listDevice := deviceList(UnusedDevices)

	// filter list with string to complete
	completions := filterCompletions(listDevice, toComplete, &directive, "No devices available for completion!")

	return completions, directive
}

// Get specific list of DPDK devices. Attached, Used and Unused
func deviceList(filter DeviceFilter) []string {
	var ports []*ethdev.Ethdev
	var key = make(map[string]bool)
	var err error
	dpdki := dpdkinfra.Get()
	list := []string{}

	// get specific lists of ethdev ports
	switch filter {
	case AllDevices:
		ports, err = dpdki.GetAttachedEthdevPorts()
	case UsedDevices:
		ports, err = dpdki.GetUsedEthdevPorts()
	case UnusedDevices:
		ports, err = dpdki.GetUnusedEthdevPorts()
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

// handle completion of sorted list of port names
func completePortList(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	var directive = cobra.ShellCompDirectiveNoFileComp // | cobra.ShellCompDirectiveNoSpace

	// get device list
	portList := portList()

	// filter list with string to complete
	completions := filterCompletions(portList, toComplete, &directive, "No Ports available for completion!")

	return completions, directive
}

// get sorted list of port names
func portList() []string {
	dpdki := dpdkinfra.Get()
	list := []string{}

	if err := dpdki.IteratePorts(func(key string, port portmngr.PortType) error {
		list = append(list, port.Name())
		return nil
	}); err != nil {
		return list
	}

	sort.Strings(list)

	return list
}

// handle completion of sorted list of unused ethdev port names
func completeUnusedPortList(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	var directive = cobra.ShellCompDirectiveNoFileComp // | cobra.ShellCompDirectiveNoSpace

	// get device list
	portList := unusedPortList()

	// filter list with string to complete
	completions := filterCompletions(portList, toComplete, &directive, "No Ports available for completion!")

	return completions, directive
}

// get sorted list of unused ethdev port names
func unusedPortList() []string {
	dpdki := dpdkinfra.Get()
	list := []string{}

	ports, err := dpdki.GetUnusedEthdevPorts()
	if err != nil {
		return list
	}

	// copy port names and sort
	for _, port := range ports {
		list = append(list, port.Name())
	}
	sort.Strings(list)

	return list
}
