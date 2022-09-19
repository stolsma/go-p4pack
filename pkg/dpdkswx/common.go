// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package dpdkswx

import "github.com/yerden/go-dpdk/common"

func err(n ...interface{}) error {
	if len(n) == 0 {
		return common.RteErrno()
	}
	return common.IntToErr(n[0])
}
