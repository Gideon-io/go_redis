package main

import "sync"

type KV struct {
	mu   sync.RWMutex
	data map[string][]byte
}

// NewKeyVal creates a new key-value store
func NewKeyVal() *KV {
	return &KV{
		data: make(map[string][]byte),
		//data: map[string][]byte{},
	}
}

// Set sets a key in the key-value store
func (kv *KV) Set(key, val []byte) error {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	kv.data[string(key)] = []byte(val)
	return nil
}

// Get gets a key from the key-value store
func (kv *KV) Get(key []byte) ([]byte, bool) {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	val, ok := kv.data[string(key)]
	return val, ok

}
