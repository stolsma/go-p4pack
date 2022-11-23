// SPDX-FileCopyrightText: 2022-present Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package logging

type registerFunc func(Logger)
type register map[string]registerFunc

var logRegister = make(register)

// Register the logging domain with a reregister callback that will be called when the logging configuration changes.
// The given callback will be called directly from this function with a current Logger based on the current
// configuration.
func Register(domain string, fn registerFunc) {
	logRegister[domain] = fn

	// logging package is not configured yet so wait until it is
	if root == nil {
		return
	}

	// initialize a logger by using the current configuration
	logger := GetLogger(domain)
	fn(logger)
}

// Will be called when the logging configuration changes
func reRegister() {
	for domain, fn := range logRegister {
		fn(GetLogger(domain))
	}
}
