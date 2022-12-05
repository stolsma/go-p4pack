// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package store

import (
	"github.com/stolsma/go-p4pack/pkg/dpdkinfra/store/kvstore"
)

type ValueInterface interface {
	kvstore.ValueInterface
	Free()
}

type Store[V ValueInterface] struct {
	*kvstore.KVStore[string, V]
}

// NewStore creates a new thread safe key-value store specifically tailored for the dpdkinfra/portmngr package
func NewStore[V ValueInterface]() *Store[V] {
	s := &Store[V]{
		KVStore: kvstore.New[string, V](),
	}
	return s
}

// Get looks up a key's value from the store and returns nil if it doesn't exist.
func (s *Store[V]) Get(key string) (value V) {
	item, ok := s.KVStore.Get(key)

	if !ok {
		return // null value for type V (with pointer type it will be nil)
	}

	return item
}

// Clear deletes all key-value pairs from the key-value store and calls the Free method on all Values.
func (s *Store[V]) Clear() {
	s.Iterate(func(key string, value V) error {
		value.Free()
		s.Delete(key)
		return nil
	})
}
