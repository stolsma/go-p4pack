// SPDX-FileCopyrightText: 2020-2022 Open Networking Foundation <info@opennetworking.org>
// SPDX-FileCopyrightText: 2022-present Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"github.com/spf13/cobra"
	"github.com/stolsma/go-p4pack/pkg/cli"
)

// GetCommand returns the root command after adding the logging service commands
func GetCommand(parents ...*cobra.Command) *cobra.Command {
	logCmd := &cobra.Command{
		Use:   "log {list}/{set/get} level [args]",
		Short: "logging api commands",
	}

	LogListCommand(logCmd)
	LogGetCommand(logCmd)
	LogSetCommand(logCmd)
	return cli.AddCommand(parents, logCmd)
}
