// SPDX-FileCopyrightText: 2022-present Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"github.com/spf13/cobra"
	"github.com/stolsma/go-p4pack/pkg/cli"
)

// GetCommand returns the root command after adding the flowtest service commands
func GetCommand(parents ...*cobra.Command) *cobra.Command {
	flowtestCmd := &cobra.Command{
		Use:   "flowtest {start/stop} test [args]",
		Short: "flowtest commands",
	}

	FlowtestStartCmd(flowtestCmd)
	FlowtestStopCmd(flowtestCmd)

	return cli.AddCommand(parents, flowtestCmd)
}
