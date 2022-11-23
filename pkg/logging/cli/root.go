// SPDX-FileCopyrightText: 2020-2022 Open Networking Foundation <info@opennetworking.org>
// SPDX-FileCopyrightText: 2022-present Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"github.com/spf13/cobra"
)

// GetCommand returns the root command after adding the logging service commands
func GetCommand(root *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "log {list}/{set/get} level [args]",
		Short: "logging api commands",
	}

	root.AddCommand(cmd)
	cmd.AddCommand(getListCommand())
	cmd.AddCommand(getGetCommand())
	cmd.AddCommand(getSetCommand())

	return root
}
