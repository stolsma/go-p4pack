package cli

import (
	"github.com/spf13/cobra"
)

func initMempool(parent *cobra.Command) {
	var mempool = &cobra.Command{
		Use:   "mempool",
		Short: "Base command for all memory pool actions",
		Run:   func(cmd *cobra.Command, args []string) {},
	}

	// implements mempool MEMPOOL0 buffer 2304 pool 32K cache 256 cpu 0
	mempool.AddCommand(&cobra.Command{
		Use:       "create [name] [buffersize] [poolsize] [cachesize] [numaid]",
		Example:   "Example: mempool create [name] [buffersize] [poolsize] [cachesize] [numaid]",
		Short:     "Create a mempool",
		Long:      `Create a mempool with name, buffersize, poolsize, cachesize and numa-id`,
		ValidArgs: []string{"name", "buffersize", "poolsize", "cachesize", "numaid"},
		Args:      cobra.ExactArgs(5),
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Printf("create %s, %s", args[0], args[1])
		},
	})

	parent.AddCommand(mempool)
}
