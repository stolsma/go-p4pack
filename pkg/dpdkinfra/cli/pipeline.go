// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/stolsma/go-p4pack/pkg/dpdkinfra"
	"github.com/stolsma/go-p4pack/pkg/dpdkswx/pipeline"
	"golang.org/x/net/context"
)

func pipelineCmd(parent *cobra.Command) *cobra.Command {
	var pipelineCmd = &cobra.Command{
		Use:     "pipeline",
		Short:   "pipeline",
		Aliases: []string{"pl"},
	}

	pipelineInfoCmd(pipelineCmd)
	pipelineStatsCmd(pipelineCmd)
	parent.AddCommand(pipelineCmd)
	return pipelineCmd
}

func pipelineInfoCmd(parent *cobra.Command) *cobra.Command {
	infoCmd := &cobra.Command{
		Use:     "info [pipeline]",
		Short:   "info",
		Aliases: []string{"i"},
		Args:    cobra.MaximumNArgs(1),
		ValidArgsFunction: ValidateArguments(
			completePipelineArg,
			AppendLastHelp(1, "This command does not take any more arguments"),
		),
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
	}

	parent.AddCommand(infoCmd)

	return infoCmd
}

func pipelineStatsCmd(parent *cobra.Command) *cobra.Command {
	var re, li, si bool
	statsCmd := &cobra.Command{
		Use:     "stats [pipeline]",
		Short:   "stats",
		Aliases: []string{"st"},
		Args:    cobra.MaximumNArgs(1),
		ValidArgsFunction: ValidateArguments(
			completePipelineArg,
			AppendLastHelp(1, "This command does not take any more arguments"),
		),
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
	statsCmd.Flags().BoolVarP(&re, "repeat", "r", false, "Continuously update statistics (every second), use CTRL-C to stop.")
	statsCmd.Flags().BoolVarP(&li, "long", "l", false, "Show all information.")
	statsCmd.Flags().BoolVarP(&si, "short", "s", true, "Show minimum information.")
	statsCmd.MarkFlagsMutuallyExclusive("long", "short")

	parent.AddCommand(statsCmd)
	return statsCmd
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

// complete a pipeline argument
func completePipelineArg(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	var directive = cobra.ShellCompDirectiveNoFileComp

	// get pipeline list
	listPl := pipelineList()

	// filter list with string to complete
	completions := filterCompletions(listPl, toComplete, &directive, "No Pipelines available for completion!")

	return completions, directive
}

// retrieve all pipeline names and return in sorted list.
func pipelineList() []string {
	dpdki := dpdkinfra.Get()

	list := []string{}
	dpdki.PipelineStore.Iterate(func(key string, value *pipeline.Pipeline) error {
		list = append(list, key)
		return nil
	})
	sort.Strings(list)

	return list
}
