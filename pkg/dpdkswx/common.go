// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package dpdkswx

/*
#include <rte_memory.h>
#include <rte_errno.h>
static int rteErrno() {
	return rte_errno;
}
*/
import "C"

import (
	"errors"
	"reflect"
	"syscall"
)

// Custom RTE induced errors.
var (
	ErrNoConfig  = errors.New("missing rte_config")
	ErrSecondary = errors.New("operation not allowed in secondary processes")
)

func errno(n int64) error {
	if n == 0 {
		return nil
	} else if n < 0 {
		n = -n
	}

	if n == int64(C.E_RTE_NO_CONFIG) {
		return ErrNoConfig
	}

	if n == int64(C.E_RTE_SECONDARY) {
		return ErrSecondary
	}

	return syscall.Errno(int(n))
}

// RteErrno returns rte_errno variable.
func RteErrno() error {
	return errno(int64(C.rteErrno()))
}

// IntToErr converts n into an 'errno' error. If n is not a signed
// integer it will panic.
func IntToErr(n interface{}) error {
	x := reflect.ValueOf(n).Int()
	return errno(x)
}

func err(n ...interface{}) error {
	if len(n) == 0 {
		return RteErrno()
	}
	return IntToErr(n[0])
}
