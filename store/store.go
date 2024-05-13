package store

import (
	"errors"
	"sync"

	"github.com/lawalbolaji/gmdb/parser"
)

/*
	supported data types
		- simple kv -> string key, string value
			supported methods:
				get key -> value
				set key value -> value

		- simple hash -> string key, hash value (string key, string value)
			supported methods:
				hset key hash_key value -> value
				hget key hash_key -> value

		- simple list -> string key, array value (strings)
			supported methods:
				lset key value -> success | failure
				add_many key values -> success | failure

				get_all key -> all values
				get_at key index -> value
*/

var kvStore = map[string]string{}
var kvStoreMtx = sync.RWMutex{}

type locker interface {
	Lock()
	Unlock()
}

type lockMeta struct {
	mode uint16 // read (0) or write (1)
	lock *sync.RWMutex
}

func (mtx *lockMeta) Lock() {
	if mtx.mode == 0 {
		mtx.lock.RLock()
		return
	}

	mtx.lock.Lock()
}

func (mtx *lockMeta) Unlock() {
	if mtx.mode == 0 {
		mtx.lock.RUnlock()
		return
	}

	mtx.lock.Unlock()
}

var commandLocks = map[string]*lockMeta{
	"GET":     {mode: 0, lock: &kvStoreMtx},
	"HGET":    {mode: 0, lock: &hashStoreMtx},
	"HGETALL": {mode: 0, lock: &hashStoreMtx},
	"SET":     {mode: 1, lock: &kvStoreMtx},
	"HSET":    {mode: 1, lock: &hashStoreMtx},
}

func GetRequiredLocks(commands [][]parser.Value) []locker {
	lockers := make([]locker, 0)

	for _, command := range commands {
		lockMeta, ok := commandLocks[command[0].Bulk]
		if ok {
			lockers = append(lockers, lockMeta)
		}
	}

	return lockers
}

func GetValFromKVStore(key string) (string, bool) {
	value, ok := kvStore[key]
	return value, ok
}

func SetValInKVStore(key string, val string) {
	kvStore[key] = val
}

var hashStore = map[string]map[string]string{}
var hashStoreMtx = sync.RWMutex{}

func SetValInHashStore(key string, hashKey string, val string) {
	if _, ok := hashStore[key]; !ok {
		hashStore[key] = map[string]string{}
	}

	hashStore[key][hashKey] = val
}

func GetValFromHashStore(key string, hashKey string) (string, bool) {
	val, ok := hashStore[key][hashKey]
	return val, ok
}

func GetAllKeysAndValFromHashStore(key string) ([]string, error) {
	hashTable, ok := hashStore[key]
	if !ok {
		return []string{}, errors.New("invalid key")
	}

	keysAndVal := make([]string, len(hashTable)*2)
	idx := 0

	for key, val := range hashTable {
		keysAndVal[idx] = key
		idx++
		keysAndVal[idx] = val
		idx++
	}

	return keysAndVal, nil
}
