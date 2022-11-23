package cli

import (
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/stolsma/go-p4pack/pkg/dpdkinfra"
	"golang.org/x/net/context"
)

const (
	ETX = 0x3 // control-C
)

func initPipeline(parent *cobra.Command) {
	var pipeline = &cobra.Command{
		Use:     "pipeline",
		Short:   "pipeline",
		Aliases: []string{"pl"},
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Printf("Pipeline")
		},
	}

	info := &cobra.Command{
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
			cmd.Printf("%s", pi)
		},
	}
	pipeline.AddCommand(info)

	var re, li, si bool
	stats := &cobra.Command{
		Use:     "stats [pipeline]",
		Short:   "stats",
		Aliases: []string{"st"},
		Args:    cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			dpdki := dpdkinfra.Get()
			ctx, cancelFn := context.WithCancel(cmd.Context())
			plName := ""

			// check specific pipeline or all
			if len(args) == 1 {
				plName = args[0]
			}

			// repeat if requested
			if !re {
				printSinglePipelineStats(cmd, dpdki, plName, 0)
			} else {
				cmd.Printf("Press CTRL-C to quit!\n")
				printRPipelineStats(ctx, cmd, dpdki, plName, re)

				// wait for CTRL-C or
				buf := make([]byte, 1)
				for {
					amount, err := cmd.InOrStdin().Read(buf)
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
				cancelFn()
				return
			}
		},
	}
	stats.Flags().BoolVarP(&re, "repeat", "r", false, "Continuously update statistics (every second), use CTRL-C to stop.")
	stats.Flags().BoolVarP(&li, "long", "l", false, "Show all information.")
	stats.Flags().BoolVarP(&si, "short", "s", true, "Show minimum information.")
	stats.MarkFlagsMutuallyExclusive("long", "short")
	pipeline.AddCommand(stats)

	parent.AddCommand(pipeline)
}

func printRPipelineStats(ctx context.Context, cmd *cobra.Command, dpdki *dpdkinfra.DpdkInfra, plName string, repeat bool) {
	go func(interval int, ctx context.Context) {
		var prevLines int
		for repeat {
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
	stats, err := dpdki.PipelineStats(plName)
	if err != nil {
		cmd.PrintErrf("Pipeline Stats err: %v\n", err)
		return 0, err
	}
	for i := 0; i <= prevLines; i++ {
		cmd.Printf("\033[A") // move the cursor up
	}
	cmd.Printf("\n%s", stats)
	return strings.Count(stats, "\n"), nil
}
