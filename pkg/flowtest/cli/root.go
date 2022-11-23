// SPDX-FileCopyrightText: 2020-2022 Open Networking Foundation <info@opennetworking.org>
// SPDX-FileCopyrightText: 2022-present Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"github.com/spf13/cobra"
)

// GetCommand returns the root command after adding the flowtest service commands
func GetCommand(root *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "flowtest {start/stop} test [args]",
		Short: "flowtest commands",
	}

	root.AddCommand(cmd)
	cmd.AddCommand(getStartCommand())
	cmd.AddCommand(getStopCommand())

	return root
}
