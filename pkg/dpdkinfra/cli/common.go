// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/stolsma/go-p4pack/pkg/dpdkinfra"
	"github.com/stolsma/go-p4pack/pkg/dpdkinfra/portmngr"
	"github.com/stolsma/go-p4pack/pkg/dpdkswx/device"
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
	AllDevices    DeviceFilter = iota + 1 // All attached DPDK devices
	UsedDevices                           // All DPDK devices with one or more (ethdev) ports created
	UnusedDevices                         // All DPDK devices with none of the related ports created
)

// handle completion of sorted list of device names with all connected ethdev ports unused (i.e. not created/bound).
func completeUnusedDeviceList(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	var directive = cobra.ShellCompDirectiveNoFileComp // | cobra.ShellCompDirectiveNoSpace

	// get device list
	listDevice := deviceList(UnusedDevices)

	// filter list with string to complete
	completions := filterCompletions(listDevice, toComplete, &directive, "No devices available for completion!")

	return completions, directive
}

// Get specific list of DPDK devices. Filtered by DeviceFilter: Attached (all), Used (i.e. with at least one port
// created) and Unused (i.e. none of the related ports created)
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
		ports, err = dpdki.GetEthdevPorts(portmngr.AllEthdevPorts)
	case UnusedDevices: // devices without created ports
		// Get a map of created device names
		var usedDevices = make(map[string]bool)
		if ports, err = dpdki.GetEthdevPorts(portmngr.AllEthdevPorts); err != nil {
			return list
		}
		for _, port := range ports {
			usedDevices[port.DevName()] = true
		}

		// copy devicenames with no created ports and filter duplicates
		if ports, err = dpdki.GetAttachedEthdevPorts(); err != nil {
			return list
		}
		for _, port := range ports {
			devName := port.DevName()
			if !usedDevices[devName] {
				list = append(list, devName)
				usedDevices[devName] = true // don't add device name multiple times when other ports from the same device
			}
		}
		sort.Strings(list)
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
	sort.Strings(list)

	return list
}

type EthdevPortFilter uint

const (
	AllEthdevPorts     EthdevPortFilter = iota + 1 // All attached raw ethdev ports
	UnusedEthdevPorts                              // All raw ethdev ports attached but not created/bound
	CreatedEthdevPorts                             // All ethdev ports created
	UnboundEthdevPorts                             // All ethdev ports created but not bound with one of its queues to a pipeline
	BoundEthdevPorts                               // All ethdev ports created and with one or more queues bound to a pipeline
)

// handle completion of sorted list of attached but unused (i.e. not created) ethdev port names
func completeUnusedEthdevPortList(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	var directive = cobra.ShellCompDirectiveNoFileComp // | cobra.ShellCompDirectiveNoSpace

	// get device list
	portList := ethdevPortList(UnusedEthdevPorts)

	// filter list with string to complete
	completions := filterCompletions(portList, toComplete, &directive, "No Ports available for completion!")

	return completions, directive
}

// get sorted list of ethdev port names filtered by given filteroption
func ethdevPortList(filter EthdevPortFilter) []string {
	var ports []*ethdev.Ethdev
	var err error
	dpdki := dpdkinfra.Get()
	list := []string{}

	switch filter {
	case UnusedEthdevPorts:
		ports, err = dpdki.GetUnusedEthdevPorts()
	case CreatedEthdevPorts:
		ports, err = dpdki.GetEthdevPorts(portmngr.AllEthdevPorts)
	case UnboundEthdevPorts:
		ports, err = dpdki.GetEthdevPorts(portmngr.UnboundEthdevPorts)
	case BoundEthdevPorts:
		ports, err = dpdki.GetEthdevPorts(portmngr.BoundEthdevPorts)
	default: // AllEthdevPorts
		ports, err = dpdki.GetAttachedEthdevPorts()
	}

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

type PortFilter uint

const (
	AllPorts     PortFilter = iota + 1 // All created ports (tap, ring, ethdev etc)
	UnboundPorts                       // All ports created but not bound with one of its queues to a pipeline
	BoundPorts                         // All ports with one or more queues bound to a pipeline
)

// handle completion of sorted list of port names
func completePortList(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	var directive = cobra.ShellCompDirectiveNoFileComp // | cobra.ShellCompDirectiveNoSpace

	// get device list
	portList := portList(AllPorts)

	// filter list with string to complete
	completions := filterCompletions(portList, toComplete, &directive, "No Ports available for completion!")

	return completions, directive
}

// get sorted list of port names filtered by given filteroption
func portList(filter PortFilter) []string {
	dpdki := dpdkinfra.Get()
	list := []string{}

	if err := dpdki.IteratePorts(func(key string, port portmngr.PortType) error {
		switch filter {
		case UnboundPorts:
			if port.IsBound() {
				return nil
			}
		case BoundPorts:
			if !port.IsBound() {
				return nil
			}
		}
		list = append(list, port.Name())
		return nil
	}); err != nil {
		return list
	}

	sort.Strings(list)

	return list
}

type QueueFilter uint

const (
	AllQueues     QueueFilter = iota + 1 // All created queues on all ports (tap, ring, ethdev etc)
	UnboundQueues                        // All created queues on all ports not bound to a pipeline port
	BoundQueues                          // All created queues on all ports bound to a pipeline port
)

// get sorted list of port + queue names filtered by given filteroption
func portQueueList(filter QueueFilter) []string {
	dpdki := dpdkinfra.Get()
	list := []string{}

	if err := dpdki.IteratePorts(func(key string, port portmngr.PortType) error {
		if err := port.IterateRxQueues(func(index uint16, q device.Queue) error {
			switch filter {
			case UnboundQueues:
				if q.PipelinePort() != device.NotBound {
					return nil
				}
			case BoundQueues:
				if q.PipelinePort() == device.NotBound {
					return nil
				}
			}
			list = append(list, fmt.Sprintf("%s:rx:%d", port.Name(), index))
			return nil
		}); err != nil {
			return err
		}

		if err := port.IterateTxQueues(func(index uint16, q device.Queue) error {
			switch filter {
			case UnboundQueues:
				if q.PipelinePort() != device.NotBound {
					return nil
				}
			case BoundQueues:
				if q.PipelinePort() == device.NotBound {
					return nil
				}
			}
			list = append(list, fmt.Sprintf("%s:tx:%d", port.Name(), index))
			return nil
		}); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return list
	}

	sort.Strings(list)

	return list
}
