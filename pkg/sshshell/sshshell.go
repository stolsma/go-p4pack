// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package sshshell

import (
	"container/ring"
	"errors"
	"fmt"
	"io"
)

const (
	ETX = 0x3  // control-C
	EOT = 0x4  // control-D
	BEL = 0x7  // beep
	BS  = 0x8  // Backspace
	TAB = 0x9  // TAB
	LF  = 0xa  // linefeed
	FF  = 0xc  // formfeed (go to top of next page)
	CR  = 0xd  // carriage return
	ESC = 0x1b // Escape
	CSI = '['  // CSI escape code following
	DEL = 0x7f // Delete
	CUU = 'A'  // Cursor Up
	CUD = 'B'  // Cursor Down
	CUF = 'C'  // Cursor Forward
	CUB = 'B'  // Cursor Back
)

type Shell struct {
	config *Config
	rw     io.ReadWriter

	history      *ring.Ring
	historyShift int
	line         []byte
	inEsc        bool
	inCsi        bool
	csiSequence  []byte
	prompt       string
	instanceName string
}

func NewShell(rw io.ReadWriter, instanceName string, config *Config) *Shell {
	return &Shell{
		rw:           rw,
		instanceName: instanceName,
		config:       config,
		history:      ring.New(useOrDefaultInt(config.HistorySize, DefaultConfigHistorySize)),
	}
}

func (s *Shell) GetReadWrite() io.ReadWriter {
	return s.rw
}

func (s *Shell) InstanceName() string {
	return s.instanceName
}

// Read blocks until the next line of input, such as a command, is available and returns that line. The error will be
// io.EOF if Control-D was pressed.
func (s *Shell) Read() (string, error) {
	buf := make([]byte, 1)

	_, err := s.rw.Write([]byte(s.prompt))
	if err != nil {
		return "", err
	}

	for {
		amount, err := s.rw.Read(buf)
		if err != nil {
			return "", err
		}

		if amount > 0 {
			ch := buf[0]
			if s.inEsc {
				switch {
				case s.inCsi:
					s.csiSequence = append(s.csiSequence, ch)
					if ch >= 0x40 || ch <= 0x7e {
						err := s.handleCsiSequence()
						if err != nil {
							return "", err
						}
					}

				case ch == CSI:
					s.inCsi = true

				default:
					fmt.Printf("Can't handle ESC sequence: %q\n", ch)
					return "", errors.New("unsupported ESC sequence")
				}
			} else {
				switch {
				case ch == ESC:
					s.inEsc = true

				case ch == DEL:
					s.unshiftHistory()
					err := s.del()
					if err != nil {
						return "", err
					}

				// line finished?
				case ch == CR:
					//					etxCh := make(chan struct{})
					line, err := s.finishLine()
					return line, err

				case ch == EOT:
					// control-D, exit
					_, err = s.rw.Write([]byte{CR, LF})
					if err != nil {
						return "", err
					}
					return "", io.EOF

				case ch == ETX:
					// cancel current input line
					s.unshiftHistory()
					_, err = s.rw.Write([]byte("\r\n" + s.prompt))
					if err != nil {
						return "", err
					}
					s.line = nil

				case ch >= ' ' && ch < '~':
					s.unshiftHistory()
					err := s.Add(ch)
					if err != nil {
						return "", err
					}

				default:
					fmt.Printf("Ignoring: %x\n", ch)
				}
			}
		}
	}
}

func (s *Shell) del() error {
	if len(s.line) > 0 {
		s.line = s.line[:len(s.line)-1]
		_, err := s.rw.Write([]byte{BS, ' ', BS})
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Shell) Add(b byte) error {
	s.line = append(s.line, b)
	_, err := s.rw.Write([]byte{b})
	return err
}

func (s *Shell) AddString(val string) error {
	if s.line == nil {
		s.line = []byte(val)
	} else {
		s.line = append(s.line, val...)
	}

	_, err := s.rw.Write([]byte(val))
	return err
}

func (s *Shell) finishLine() (string, error) {
	_, err := s.rw.Write([]byte{CR, LF})
	if err != nil {
		return "", err
	}

	line := string(s.line)
	if s.historyShift == 0 && line != "" {
		s.history.Value = line
		s.history = s.history.Next()
	} else {
		// pressed enter on a history item, so don't re-save it
		s.unshiftHistory()
	}
	s.line = nil
	return line, nil
}

func (s *Shell) outputShiftedHistory() error {
	err := s.eraseCurrent()
	if err != nil {
		return err
	}
	err = s.AddString(s.history.Value.(string))
	if err != nil {
		return err
	}
	return err
}

func (s *Shell) SetPrompt(prompt string) {
	s.prompt = prompt
}

func (s *Shell) handleCsiSequence() error {
	s.inEsc = false
	s.inCsi = false

	switch s.csiSequence[len(s.csiSequence)-1] {
	case CUU:
		return s.cursorUp()
	case CUD:
		return s.cursorDown()
	default:
		fmt.Printf("Ignoring CSI: %s\n", s.csiSequence)
		return nil
	}
}

func (s *Shell) OutputLine(line string) error {
	_, err := s.rw.Write([]byte(line + "\r\n"))
	return err
}

func (s *Shell) unshiftHistory() {
	s.history = s.history.Move(s.historyShift)
	s.historyShift = 0
}

func (s *Shell) cursorUp() error {
	prev := s.history.Prev()
	if prev.Value == nil {
		err := s.Bell()
		if err != nil {
			return err
		}
		return nil
	}

	s.history = prev
	s.historyShift++

	return s.outputShiftedHistory()
}

func (s *Shell) cursorDown() error {
	if s.historyShift <= 0 {
		err := s.Bell()
		if err != nil {
			return err
		}
		return nil
	}

	s.history = s.history.Next()
	s.historyShift--

	if s.historyShift == 0 {
		return s.eraseCurrent()
	}

	return s.outputShiftedHistory()
}

func (s *Shell) Bell() error {
	_, err := s.rw.Write([]byte{BEL})
	return err
}

func (s *Shell) eraseCurrent() error {
	amount := len(s.line)
	s.line = nil

	err := s.left(amount)
	if err != nil {
		return err
	}

	for i := 0; i < amount; i++ {
		_, err := s.rw.Write([]byte{' '})
		if err != nil {
			return err
		}
	}

	err = s.left(amount)
	if err != nil {
		return err
	}

	return nil
}

func (s *Shell) left(n int) error {
	for ; n > 0; n-- {
		_, err := s.rw.Write([]byte{BS})
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Shell) Refresh() error {
	if err := s.OutputLine(""); err != nil {
		return err
	}
	_, err := s.rw.Write([]byte(fmt.Sprintf("\r\n%s%s", s.prompt, string(s.line))))
	if err != nil {
		return err
	}
	return nil
}
