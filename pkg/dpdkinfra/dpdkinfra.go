// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package dpdkinfra

import (
	"log"
)

func DpdkInfraInit(dpdkArgs []string) error {
	numArgs, status := EalInit(dpdkArgs)
	if status != nil {
		return status
	}
	log.Printf("EAL init ok! Num Args:%d", numArgs)

	if status := ObjInit(); status != nil {
		return status
	}
	log.Println("ObjInit ok!")

	if status := ThreadInit(); status != nil {
		return status
	}
	log.Println("ThreadInit ok!")

	if status := MainThreadInit(); status != nil {
		return status
	}
	log.Println("MainThreadInit ok!")

	return nil
}
