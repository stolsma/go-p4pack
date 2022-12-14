// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"strconv"

	"github.com/spf13/cobra"
	"github.com/stolsma/go-p4pack/pkg/dpdkinfra"
	"github.com/stolsma/go-p4pack/pkg/dpdkswx/pktmbuf"
)

func initPktmbuf(parent *cobra.Command) {
	var pktmbufCmd = &cobra.Command{
		Use:   "pktmbuf",
		Short: "Base command for all pktmbuf actions",
	}

	// implements mempool MEMPOOL0 buffer 2304 pool 32K cache 256 cpu 0
	pktmbufCmd.AddCommand(&cobra.Command{
		Use:     "create [name] [buffersize] [poolsize] [cachesize] [numaid]",
		Example: "Example: pktmbuf create [name] [buffersize] [poolsize] [cachesize] [numaid]",
		Short:   "Create a pktmbuf",
		Long:    `Create a pktmbuf with name, buffersize, poolsize, cachesize and numa-id`,
		Args:    cobra.ExactArgs(5),
		Run: func(cmd *cobra.Command, args []string) {
			dpdki := dpdkinfra.Get()

			// parse arguments
			var name = args[0]
			bufferSize, err := strconv.ParseUint(args[1], 10, 32)
			if err != nil {
				cmd.PrintErrf("Buffersize parse err: %v\n", err)
				return
			}
			poolSize, err := strconv.ParseUint(args[2], 10, 32)
			if err != nil {
				cmd.PrintErrf("Poolsize parse err: %v\n", err)
				return
			}
			cacheSize, err := strconv.ParseUint(args[3], 10, 32)
			if err != nil {
				cmd.PrintErrf("Cachesize parse err: %v\n", err)
				return
			}
			numaID, err := strconv.ParseInt(args[4], 10, 32)
			if err != nil {
				cmd.PrintErrf("NumaID parse err: %v\n", err)
				return
			}

			// create
			_, err = dpdki.PktmbufCreate(name, uint(bufferSize), uint32(poolSize), uint32(cacheSize), int(numaID))
			if err != nil {
				cmd.PrintErrf("Pktmbuf create err: %v\n", err)
			} else {
				cmd.Printf("Pktmbuf %s created!\n", name)
			}
		},
	})

	pktmbufCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List all created Pktmbuf",
		Run: func(cmd *cobra.Command, args []string) {
			dpdki := dpdkinfra.Get()

			err := dpdki.PktmbufStore.Iterate(func(key string, value *pktmbuf.Pktmbuf) error {
				cmd.Println(key)
				return nil
			})

			if err != nil {
				cmd.PrintErrf("List err: %v\n", err)
			}
		},
	})

	parent.AddCommand(pktmbufCmd)
}
