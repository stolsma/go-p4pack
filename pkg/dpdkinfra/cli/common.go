// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"io"
	"strings"
)

const (
	ETX = 0x3 // control-C
)

// wait for CTRL-C
func waitForCtrlC(input io.Reader) {
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

func indent(chars string, orig string) string {
	var last bool
	interm := strings.Split(orig, "\n")
	if interm[len(interm)-1] == "" {
		interm = interm[:len(interm)-1]
		last = true
	}
	res := chars + strings.Join(interm, "\n"+chars)
	if last {
		return res + "\n"
	}
	return res
}
