package main

import "sync"

type KV struct {
	mu   sync.RWMutex
	data map[string][]byte
}

func NewKeyVal() *KV {
	return &KV{
		data: make(map[string][]byte),
		//data: map[string][]byte{},
	}
}

func (kv *KV) Set(key, val []byte) error {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	kv.data[string(key)] = []byte(val)
	return nil
}

func (kv *KV) Get(key []byte) ([]byte, bool) {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	val, ok := kv.data[string(key)]
	return val, ok

}
