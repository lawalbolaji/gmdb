package store

import (
	"sync"
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

var regularStore = map[string]string{}
var regStoreMtx = sync.RWMutex{}

func GetValFromKVStore(key string) (string, bool) {
	regStoreMtx.RLock()
	defer regStoreMtx.RUnlock()

	value, ok := regularStore[key]
	return value, ok
}

func SetValInKVStore(key string, val string) {
	regStoreMtx.Lock()
	defer regStoreMtx.Unlock()

	regularStore[key] = val
}

var hashStore = map[string]map[string]string{}
var hashStoreMtx = sync.RWMutex{}

func SetValInHashStore(key string, hashKey string, val string) {
	hashStoreMtx.Lock()
	defer hashStoreMtx.Unlock()

	if _, ok := hashStore[key]; !ok {
		hashStore[key] = map[string]string{}
	}

	hashStore[key][hashKey] = val
}

func GetValFromHashStore(key string, hashKey string) (string, bool) {
	hashStoreMtx.RLock()
	defer hashStoreMtx.RUnlock()

	val, ok := hashStore[key][hashKey]
	return val, ok
}
