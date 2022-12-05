// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package kvstore

import (
	"errors"
	"sort"
	"sync"

	"golang.org/x/exp/constraints"
)

// ValueInterface is a common value interface.
type ValueInterface interface {
}

// Interface is a common key-value store interface.
type Interface[K constraints.Ordered, V ValueInterface] interface {
	// Get looks up a key's value from the key-value store. Ok is false if not found.
	Get(key K) (value *V, ok bool)
	// Set sets a value to the key-value store with key, replacing any existing value.
	Set(key K, val *V)
	// Keys returns the keys of the key-value store. The order is relied on algorithms.
	Keys() []K
	// Delete deletes the item with provided key from the key-value store.
	Delete(key K)
	// Contains reports whether the given key is stored within the key-value store.
	Contains(key K) bool
	// Iterate iterates over all the items in the key-value store and calls given function with key, value and options.
	Iterate(fn func(key K, value *V) error) error
}

// Item is an kvstore item
type Item[K constraints.Ordered, V ValueInterface] struct {
	Key   K
	Value V
}

// newItem creates a new item with specified options.
func newItem[K constraints.Ordered, V ValueInterface](key K, val V) *Item[K, V] {
	return &Item[K, V]{
		Key:   key,
		Value: val,
	}
}

// KVStore is a thread safe key-value store.
type KVStore[K constraints.Ordered, V ValueInterface] struct {
	s map[K]*Item[K, V]
	// mu is used to do lock in some method process.
	mu sync.Mutex
}

// New creates a new thread safe key-value store.
func New[K constraints.Ordered, V ValueInterface]() *KVStore[K, V] {
	cache := &KVStore[K, V]{
		s: make(map[K]*Item[K, V], 0),
	}
	return cache
}

// Get looks up a key's value from the key-value store.
func (c *KVStore[K, V]) Get(key K) (value V, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	item, ok := c.s[key]

	if !ok {
		return
	}

	return item.Value, true
}

// Set sets a value to the key-value store with key. Replacing any existing value.
func (c *KVStore[K, V]) Set(key K, val V) {
	c.mu.Lock()
	defer c.mu.Unlock()

	item := newItem(key, val)
	c.s[key] = item
}

// Keys returns the keys of the key-value store. The order is sorted from low to high (i.e. a-Z).
func (c *KVStore[K, V]) Keys() []K {
	c.mu.Lock()
	defer c.mu.Unlock()

	ret := make([]K, 0, len(c.s))
	for key := range c.s {
		ret = append(ret, key)
	}
	sort.Slice(ret, func(i, j int) bool {
		return c.s[ret[i]].Key == c.s[ret[j]].Key
	})

	return ret
}

// Delete deletes the item identified with the given key from the key-value store.
func (c *KVStore[K, V]) Delete(key K) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.s, key)
}

// Contains reports whether the given key is stored within the key-value store.
func (c *KVStore[K, V]) Contains(key K) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	_, ok := c.s[key]
	return ok
}

// Iterate iterates over all the items in the key-value store and calls given function with key, value and options.
func (c *KVStore[K, V]) Iterate(fn func(key K, value V) error) error {
	if fn != nil {
		for k, v := range c.s {
			if err := fn(k, v.Value); err != nil {
				return err
			}
		}
	} else {
		return errors.New("no function to call")
	}
	return nil
}
