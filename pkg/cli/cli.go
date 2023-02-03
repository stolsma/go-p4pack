// Copyright 2023 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"io"
	"strings"

	"github.com/spf13/cobra"
)

const (
	ETX = 0x3 // control-C
)

// wait for CTRL-C
func WaitForCtrlC(input io.Reader) {
	buf := make([]byte, 1)
	for {
		amount, err := input.Read(buf)
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
}

// add cobra command to given parent(s) and returns given command
func AddCommand(parents []*cobra.Command, command *cobra.Command) *cobra.Command {
	if len(parents) > 0 {
		parents[0].AddCommand(command)
	}

	return command
}

type ValidateFn = func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective)

// execute completion commands per argument
func ValidateArguments(fns ...ValidateFn) ValidateFn {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		var completions []string
		var directive cobra.ShellCompDirective
		var fnLen = len(fns)
		var argLen = len(args)

		// call completion function related to the current argument position
		if argLen <= fnLen-2 {
			completions, directive = fns[argLen](cmd, args, toComplete)
		}

		// if last real argument has no completion or if pos of current argument > max arguments call catch-all function
		if argLen >= fnLen-2 && len(completions) == 0 {
			comp, direct := fns[fnLen-1](cmd, args, toComplete)
			if len(comp) > 0 {
				completions = comp
				directive = direct
			}
		}

		return completions, directive
	}
}

// add active help text for argument, and show if argument is empty
func AppendHelp(helpTxt string) ValidateFn {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		var completions []string
		directive := cobra.ShellCompDirectiveNoFileComp

		// help needed
		if toComplete == "" {
			completions = cobra.AppendActiveHelp(completions, helpTxt)
			directive |= cobra.ShellCompDirectiveNoSpace
		}

		return completions, directive
	}
}

// add active help text shown when last valid argument is already entered
func AppendLastHelp(total int, helpTxt string) ValidateFn {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		var completions []string
		directive := cobra.ShellCompDirectiveNoFileComp

		// help needed
		if len(args) == total-1 && toComplete != "" || len(args) > total-1 {
			completions = cobra.AppendActiveHelp(completions, helpTxt)
			directive |= cobra.ShellCompDirectiveNoSpace
		}

		return completions, directive
	}
}

// filter all strings starting with string to complete
func FilterCompletions(list []string, toComplete string, directive *cobra.ShellCompDirective, help string) []string {
	var completions []string

	// always no space
	*directive |= cobra.ShellCompDirectiveNoSpace

	// filter list
	for _, name := range list {
		if strings.HasPrefix(name, toComplete) {
			completions = append(completions, name)
		}
	}

	// if none filtered, select all strings
	if len(completions) == 0 {
		completions = list
	}

	// if only one answer left and that is same as toComplete return nothing
	if len(completions) == 1 && completions[0] == toComplete {
		completions = []string{}
		*directive ^= cobra.ShellCompDirectiveNoSpace
	} else if len(completions) == 0 {
		completions = cobra.AppendActiveHelp(completions, help)
	}

	return completions
}
