package cli

import (
	"sort"

	"github.com/spf13/cobra"
	"github.com/stolsma/go-p4pack/pkg/cli"
	"github.com/stolsma/go-p4pack/pkg/logging"
)

func LogListCommand(parents ...*cobra.Command) *cobra.Command {
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "Shows all logger domains",
		Run:   runListCommand,
	}

	return cli.AddCommand(parents, listCmd)
}

func runListCommand(cmd *cobra.Command, args []string) {
	var rootString = "root (default, not changeable)"

	// get the loggers operational configuration
	list := logging.GetLoggerDataList()

	// sort the returned domains
	keys := make([]string, 0, len(list))
	var maxLen = len(rootString)
	for k := range list {
		keys = append(keys, k)
		if l := len(k); l > maxLen {
			maxLen = l
		}
	}
	sort.Strings(keys)

	// print the info on domain sorted order
	cmd.Printf("Current defined loging domains and loglevel:\n")
	cmd.Printf("  %-*s  LOG LEVEL\n", maxLen, "DOMAIN")
	for _, domain := range keys {
		if domain == "root" {
			cmd.Printf("  %-*s  %s\n", maxLen, rootString, list[domain].Level)
		} else {
			cmd.Printf("  %-*s  %s\n", maxLen, domain, list[domain].Level)
		}
	}
}
