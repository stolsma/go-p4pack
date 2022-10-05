// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package sshshell

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"

	"github.com/gliderlabs/ssh"
)

// SshServer manages acceptance of and authenticating SSH connections and delegating input to a Handler for each
// session instantiated by the given HandlerFactory.
type SSHServer struct {
	*Config
	HandlerFactory
	cancelFn context.CancelFunc
}

type HandlerFactory func(s *Shell) Handler

type Handler interface {
	// HandleLine is called with the next line that was consumed by the SSH shell. Typically this is due the user typing
	// a command string. If an error is returned, then the error is reported back to the SSH client and the SSH session
	// is closed.
	HandleLine(ctx context.Context, line string) error

	// HandleEOF is called when the user types Control-D. If an error is returned, then the error is reported back to the
	// SSH client and the SSH session is closed.
	HandleEOF() error
}

// Listen will block listening for new SSH connections and serving those with a new instance of Shell/Handler.
func (s *SSHServer) Listen(ctx context.Context) error {
	config := s.Config
	if config == nil {
		config = &Config{}
	}

	// the basis context for the SSH server with cancel function
	srvCtx, cancelFn := context.WithCancel(ctx)
	s.cancelFn = cancelFn

	// get the handler factory function
	handlerFactory := s.HandlerFactory
	if handlerFactory == nil {
		return errors.New("no ssh server handler function present")
	}

	auth := NewAuth(config)
	hostKeyResolver := NewHostKeyResolver(config)
	bind := useOrDefaultString(config.Bind, DefaultBind)

	log.Printf("Accepting SSH connections at %s", bind)
	return ssh.ListenAndServe(bind,
		func(session ssh.Session) {
			// the context for this session
			sessionCtx, cancelSession := context.WithCancel(srvCtx)

			sesName := fmt.Sprintf("%s@%s", session.User(), session.RemoteAddr())
			log.Printf("I: New session for user=%s from=%s\n", session.User(), session.RemoteAddr())
			shell := NewShell(session, sesName, config)
			shell.SetPrompt("> ")

			handler := handlerFactory(shell)

			// handle incoming data
			for {
				select {
				case <-sessionCtx.Done(): // Quit session when it needs to stop
					endSession(session, shell, "This server will be closed!", 0)
					cancelSession()
					return
				default:
					// read line from shell
					line, err := shell.Read()

					if err != nil {
						if err == io.EOF {
							err = handler.HandleEOF()
							if err != nil && err != io.EOF {
								endSessionWithError(session, shell, err)
							}
						} else {
							endSessionWithError(session, shell, err)
						}
						cancelSession()
						return
					}

					// handle the line read, if returned error is io.EOF then stop silent else return with the error
					err = handler.HandleLine(sessionCtx, line)
					if err != nil {
						if err != io.EOF {
							endSessionWithError(session, shell, err)
						}
						cancelSession()
						return
					}
				}
			}
		},
		ssh.PasswordAuth(auth.PasswordHandler),
		hostKeyResolver.ResolveOption(),
	)
}

// Stop the SSH server and any open session!
func (s *SSHServer) Quit() {
	s.cancelFn()
}

func endSession(s ssh.Session, shell *Shell, msg string, code int) {
	_ = shell.OutputLine("")
	_ = shell.OutputLine(msg)
	_ = s.Exit(code)
}

func endSessionWithError(s ssh.Session, shell *Shell, err error) {
	endSession(s, shell, err.Error(), 1)
}
