// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/stolsma/go-p4pack/pkg/dpdkinfra"
	"golang.org/x/net/context"
)

func initPipeline(parent *cobra.Command) {
	var pipeline = &cobra.Command{
		Use:     "pipeline",
		Short:   "pipeline",
		Aliases: []string{"pl"},
	}

	initPlInfo(pipeline)
	initPlStats(pipeline)
	parent.AddCommand(pipeline)
}

func initPlInfo(parent *cobra.Command) {
	parent.AddCommand(&cobra.Command{
		Use:     "info [pipeline]",
		Short:   "info",
		Aliases: []string{"i"},
		Args:    cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			dpdki := dpdkinfra.Get()
			plName := ""

			// check specific pipeline or all
			if len(args) == 1 {
				plName = args[0]
			}

			pi, err := dpdki.PipelineInfo(plName)
			if err != nil {
				cmd.PrintErrf("Pipeline Info err: %v\n", err)
				return
			}

			for plName, plInfo := range pi {
				cmd.Printf("%s: \n", plName)
				cmd.Printf(plInfo.String())
			}
		},
	})
}

func initPlStats(parent *cobra.Command) {
	var re, li, si bool
	stats := &cobra.Command{
		Use:     "stats [pipeline]",
		Short:   "stats",
		Aliases: []string{"st"},
		Args:    cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			dpdki := dpdkinfra.Get()
			plName := ""

			// check specific pipeline or all
			if len(args) == 1 {
				plName = args[0]
			}

			// repeat output if requested
			if !re {
				printSinglePipelineStats(cmd, dpdki, plName, 0)
			} else {
				ctx, cancelFn := context.WithCancel(cmd.Context())

				cmd.Printf("Press CTRL-C to quit!\n")
				printRepeatedPipelineStats(ctx, cmd, dpdki, plName)

				// wait for CTRL-C and then cancel output
				waitForCtrlC(cmd.InOrStdin())
				cancelFn()
				return
			}
		},
	}
	stats.Flags().BoolVarP(&re, "repeat", "r", false, "Continuously update statistics (every second), use CTRL-C to stop.")
	stats.Flags().BoolVarP(&li, "long", "l", false, "Show all information.")
	stats.Flags().BoolVarP(&si, "short", "s", true, "Show minimum information.")
	stats.MarkFlagsMutuallyExclusive("long", "short")
	parent.AddCommand(stats)
}

func printRepeatedPipelineStats(ctx context.Context, cmd *cobra.Command, dpdki *dpdkinfra.DpdkInfra, plName string) {
	go func(interval int, ctx context.Context) {
		var prevLines int
		for {
			timeout := time.Duration(interval) * time.Second
			tCtx, tCancel := context.WithTimeout(ctx, timeout)
			select {
			case <-tCtx.Done():
				var err error
				prevLines, err = printSinglePipelineStats(cmd, dpdki, plName, prevLines)
				if err != nil {
					return
				}
			case <-ctx.Done():
				tCancel()
				return
			}
		}
	}(1, ctx)
}

func printSinglePipelineStats(cmd *cobra.Command, dpdki *dpdkinfra.DpdkInfra, plName string, prevLines int) (int, error) {
	plStats, err := dpdki.PipelineStats(plName)
	if err != nil {
		cmd.PrintErrf("Pipeline Stats err: %v\n", err)
		return 0, err
	}

	for i := 0; i <= prevLines; i++ {
		cmd.Printf("\033[A") // move the cursor up
	}

	var stats string
	for plName, plStat := range plStats {
		stats += fmt.Sprintf("%s:\n%v\n", plName, plStat.String())
	}

	cmd.Printf("\n%s", stats)
	return strings.Count(stats, "\n"), nil
}
