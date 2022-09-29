// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package sshshell

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"log"
	"os"

	"github.com/gliderlabs/ssh"
	gossh "golang.org/x/crypto/ssh"
)

type HostKeyResolver struct {
	hostKeyFile string
}

func NewHostKeyResolver(config *Config) *HostKeyResolver {
	return &HostKeyResolver{
		hostKeyFile: config.HostKeyFile,
	}
}

func (r *HostKeyResolver) Resolve() string {
	if r.hostKeyFile == "" {
		return ""
	}

	_, err := os.Stat(r.hostKeyFile)
	if err != nil {
		if os.IsNotExist(err) {
			err := r.createHostKeyFile()
			if err != nil {
				log.Fatal(err)
			}
		} else {
			log.Fatal(err)
		}
	}

	return r.hostKeyFile
}

func (r *HostKeyResolver) createHostKeyFile() error {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	publicKey, err := gossh.NewPublicKey(key.Public())
	if err != nil {
		return err
	}
	fingerprint := gossh.FingerprintSHA256(publicKey)
	log.Printf("I: Generating host key with fingerprint %s to %s\n", fingerprint, r.hostKeyFile)
	file, err := os.Create(r.hostKeyFile)
	if err != nil {
		return err
	}
	//noinspection GoUnhandledErrorResult
	defer file.Close()

	var keyBlock = &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}

	err = pem.Encode(file, keyBlock)
	return err
}

func (r *HostKeyResolver) ResolveOption() ssh.Option {
	if r.hostKeyFile == "" {
		return func(s *ssh.Server) error {
			return nil
		}
	} else {
		return ssh.HostKeyFile(r.Resolve())
	}
}
