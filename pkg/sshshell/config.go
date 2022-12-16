// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package sshshell

type User struct {
	Password string `json:"password"`
}

type Config struct {
	HistorySize int             `json:"historysize"`
	HostKeyFile string          `json:"hostkeyfile" mapstructure:"host-key-file"`
	Users       map[string]User `json:"users"`
	Bind        string          `json:"bind"`
}

const (
	DefaultConfigHistorySize = 100
	DefaultBind              = ":22"
)

func useOrDefaultInt(value int, defaultValue int) int {
	if value == 0 {
		return defaultValue
	}
	return value
}

func useOrDefaultString(value string, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return value
}
