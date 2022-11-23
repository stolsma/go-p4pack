// SPDX-FileCopyrightText: 2020-2022 Open Networking Foundation <info@opennetworking.org>
// SPDX-FileCopyrightText: 2022-present Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package logging

import (
	"strings"

	zp "go.uber.org/zap"
	zc "go.uber.org/zap/zapcore"
)

// Level :
type Level int

const (
	// DebugLevel logs a message at debug level
	DebugLevel Level = iota
	// InfoLevel logs a message at info level
	InfoLevel
	// WarnLevel logs a message at warning level
	WarnLevel
	// ErrorLevel logs a message at error level
	ErrorLevel
	// FatalLevel logs a message, then calls os.Exit(1).
	FatalLevel
	// PanicLevel logs a message, then panics.
	PanicLevel
	// DPanicLevel logs at PanicLevel; otherwise, it logs at ErrorLevel
	DPanicLevel
	// Last level in the list
	LastLevel

	// EmptyLevel :
	EmptyLevel = InfoLevel
)

type LevelStringsType []string

func (ls LevelStringsType) String() string {
	return strings.Join(ls, ", ")
}

var LevelStrings = LevelStringsType{
	DebugLevel:  "DEBUG",
	InfoLevel:   "INFO",
	WarnLevel:   "WARN",
	ErrorLevel:  "ERROR",
	FatalLevel:  "FATAL",
	PanicLevel:  "PANIC",
	DPanicLevel: "DPANIC",
}

// String :
func (l Level) String() string {
	if l >= LastLevel || l < 0 {
		return ""
	}
	return LevelStrings[l]
}

func levelToAtomicLevel(l Level) zp.AtomicLevel {
	switch l {
	case DebugLevel:
		return zp.NewAtomicLevelAt(zc.DebugLevel)
	case InfoLevel:
		return zp.NewAtomicLevelAt(zc.InfoLevel)
	case WarnLevel:
		return zp.NewAtomicLevelAt(zc.WarnLevel)
	case ErrorLevel:
		return zp.NewAtomicLevelAt(zc.ErrorLevel)
	case FatalLevel:
		return zp.NewAtomicLevelAt(zc.FatalLevel)
	case PanicLevel:
		return zp.NewAtomicLevelAt(zc.PanicLevel)
	case DPanicLevel:
		return zp.NewAtomicLevelAt(zc.DPanicLevel)
	}
	return zp.NewAtomicLevelAt(zc.ErrorLevel)
}

func levelStringToLevel(l string) Level {
	lvl := LevelString2Level(l)
	if lvl == LastLevel {
		return ErrorLevel
	}
	return lvl
}

func LevelString2Level(l string) Level {
	switch strings.ToUpper(l) {
	case DebugLevel.String():
		return DebugLevel
	case InfoLevel.String():
		return InfoLevel
	case WarnLevel.String():
		return WarnLevel
	case ErrorLevel.String():
		return ErrorLevel
	case FatalLevel.String():
		return FatalLevel
	case PanicLevel.String():
		return PanicLevel
	case DPanicLevel.String():
		return DPanicLevel
	}
	return LastLevel
}
