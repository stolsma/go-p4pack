// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"
	"github.com/stolsma/go-p4pack/pkg/cli"
	"github.com/stolsma/go-p4pack/pkg/dpdkinfra"
	"github.com/stolsma/go-p4pack/pkg/dpdkinfra/portmngr"
	"github.com/stolsma/go-p4pack/pkg/dpdkswx/device"
	"github.com/stolsma/go-p4pack/pkg/dpdkswx/ethdev"
	"github.com/stolsma/go-p4pack/pkg/dpdkswx/pktmbuf"
)

// complete a pktmbuf argument
func completePktmbufArg(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	var directive = cobra.ShellCompDirectiveNoFileComp // | cobra.ShellCompDirectiveNoSpace

	// get pktmbuf list
	listPktmbuf := pktmbufList()

	// filter list with string to complete
	completions := cli.FilterCompletions(listPktmbuf, toComplete, &directive, "No Pktmbufs available for completion!")

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
	completions := cli.FilterCompletions(listDevice, toComplete, &directive, "No devices available for completion!")

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
	completions := cli.FilterCompletions(portList, toComplete, &directive, "No Ports available for completion!")

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
	completions := cli.FilterCompletions(portList, toComplete, &directive, "No Ports available for completion!")

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
