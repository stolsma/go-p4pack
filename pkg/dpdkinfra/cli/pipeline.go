// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/stolsma/go-p4pack/pkg/cli"
	"github.com/stolsma/go-p4pack/pkg/dpdkinfra"
	"github.com/stolsma/go-p4pack/pkg/dpdkswx/pipeline"
	"golang.org/x/net/context"
)

func PipelineCmd(parents ...*cobra.Command) *cobra.Command {
	var pipelineCmd = &cobra.Command{
		Use:     "pipeline",
		Short:   "pipeline",
		Aliases: []string{"pl"},
	}

	PipelineBindCmd(pipelineCmd)
	PipelineInfoCmd(pipelineCmd)
	PipelineStatsCmd(pipelineCmd)
	return cli.AddCommand(parents, pipelineCmd)
}

func PipelineBindCmd(parents ...*cobra.Command) *cobra.Command {
	bindCmd := &cobra.Command{
		Use:     "bind [interface:{rx|tx}:queue#] [pipeline] [pipeline port] [burstsize]",
		Short:   "b",
		Aliases: []string{"bin"},
		Args:    cobra.RangeArgs(2, 4),
		ValidArgsFunction: cli.ValidateArguments(
			completeUnboundQueueArg,
			completeBuildNotEnabledPipelineArg,
			cli.AppendHelp("You must specify the pipeline port for the queue you are binding"),
			cli.AppendHelp("You must specify the burstsize for the queue you are binding"),
			cli.AppendLastHelp(4, "This command does not take any more arguments"),
		),
		Run: func(cmd *cobra.Command, args []string) {
			/* TODO Handle bind command
			dpdki := dpdkinfra.Get()
			plName := ""

			// check specific pipeline or all
			if len(args) == 2 {
				plName = args[1]
			}
			*/
		},
	}

	return cli.AddCommand(parents, bindCmd)
}

func PipelineInfoCmd(parents ...*cobra.Command) *cobra.Command {
	infoCmd := &cobra.Command{
		Use:     "info [pipeline]",
		Short:   "i",
		Aliases: []string{"i"},
		Args:    cobra.MaximumNArgs(1),
		ValidArgsFunction: cli.ValidateArguments(
			completePipelineArg,
			cli.AppendLastHelp(1, "This command does not take any more arguments"),
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

	return cli.AddCommand(parents, infoCmd)
}

func PipelineStatsCmd(parents ...*cobra.Command) *cobra.Command {
	var re, li, si bool
	statsCmd := &cobra.Command{
		Use:     "stats [pipeline]",
		Short:   "st",
		Aliases: []string{"st"},
		Args:    cobra.MaximumNArgs(1),
		ValidArgsFunction: cli.ValidateArguments(
			completePipelineArg,
			cli.AppendLastHelp(1, "This command does not take any more arguments"),
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
				cli.WaitForCtrlC(cmd.InOrStdin())
				cancelFn()
				return
			}
		},
	}
	statsCmd.Flags().BoolVarP(&re, "repeat", "r", false, "Continuously update statistics (every second), use CTRL-C to stop.")
	statsCmd.Flags().BoolVarP(&li, "long", "l", false, "Show all information.")
	statsCmd.Flags().BoolVarP(&si, "short", "s", true, "Show minimum information.")
	statsCmd.MarkFlagsMutuallyExclusive("long", "short")
	return cli.AddCommand(parents, statsCmd)
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

// complete an all pipeline argument
func completePipelineArg(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	var directive = cobra.ShellCompDirectiveNoFileComp

	// get pipeline list
	listPl := pipelineList(AllPipelines)

	// filter list with string to complete
	completions := cli.FilterCompletions(listPl, toComplete, &directive, "No Pipelines available for completion!")

	return completions, directive
}

// complete an BuildNotEnabledPipelines argument
func completeBuildNotEnabledPipelineArg(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	var directive = cobra.ShellCompDirectiveNoFileComp

	// get BuildNotEnabledPipelines list
	listPl := pipelineList(BuildNotEnabledPipelines)

	// filter list with string to complete
	completions := cli.FilterCompletions(listPl, toComplete, &directive, "No Pipelines available for completion!")

	return completions, directive
}

// complete a queue argument
func completeUnboundQueueArg(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	var directive = cobra.ShellCompDirectiveNoFileComp

	// get queue list
	qList := portQueueList(UnboundQueues)

	// filter list with string to complete
	completions := cli.FilterCompletions(qList, toComplete, &directive, "No queues available for completion!")

	return completions, directive
}

type PipelineFilter uint

const (
	AllPipelines PipelineFilter = iota + 1
	BuildPipelines
	BuildNotEnabledPipelines
	EnabledPipelines
)

// retrieve all pipeline names and return in sorted list.
func pipelineList(filter PipelineFilter) []string {
	dpdki := dpdkinfra.Get()

	list := []string{}
	dpdki.PipelineStore.Iterate(func(key string, pl *pipeline.Pipeline) error {
		switch filter {
		case BuildPipelines:
			if !pl.IsBuild() {
				return nil
			}
		case BuildNotEnabledPipelines:
			if !pl.IsBuild() || pl.IsEnabled() {
				return nil
			}
		case EnabledPipelines:
			if !pl.IsEnabled() {
				return nil
			}
		}

		list = append(list, key)
		return nil
	})
	sort.Strings(list)

	return list
}
