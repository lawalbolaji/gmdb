package commands

import (
	"gmdb/parser"
	"sync"
)

func ping(args []parser.Value) parser.Value {
	if len(args) == 0 {
		return parser.Value{Typ: parser.SIMPLE_STRING, Str: "PONG"}
	}

	return parser.Value{Typ: parser.BULK_STRING, Bulk: args[0].Bulk}
}

var Handlers = map[string]func([]parser.Value) parser.Value{
	"PING": ping,
	"SET":  set,
	"GET":  get,
	"HGET": hget,
	"HSET": hset,
}

var regularStore = map[string]string{}
var regStoreMtx = sync.RWMutex{}

func set(args []parser.Value) parser.Value {
	if len(args) != 2 {
		return parser.Value{Typ: parser.SIMPLE_ERROR, Str: "ERR wrong number of arguments for 'set' command"}
	}

	key := args[0].Bulk
	val := args[1].Bulk

	regStoreMtx.Lock()
	regularStore[key] = val
	regStoreMtx.Unlock()

	return parser.Value{Typ: parser.SIMPLE_STRING, Str: "OK"}
}

func get(args []parser.Value) parser.Value {
	if len(args) != 1 {
		return parser.Value{Typ: parser.SIMPLE_ERROR, Str: "ERR wrong number of arguments for 'get' command"}
	}

	key := args[0].Bulk

	regStoreMtx.RLock()
	value, ok := regularStore[key]
	regStoreMtx.RUnlock()

	if !ok {
		return parser.Value{Typ: parser.NULL}
	}
	return parser.Value{Typ: parser.BULK_STRING, Bulk: value}
}

var hashStore = map[string]map[string]string{}
var hashStoreMtx = sync.RWMutex{}

func hset(args []parser.Value) parser.Value {
	if len(args) != 3 { // hashmap, key, val
		return parser.Value{Typ: parser.SIMPLE_ERROR, Str: "ERR wrong number of arguments for 'hset' command"}
	}

	hash := args[0].Bulk
	key := args[1].Bulk
	val := args[2].Bulk

	hashStoreMtx.Lock()
	if _, ok := hashStore[hash]; !ok {
		hashStore[hash] = map[string]string{}
	}
	hashStore[hash][key] = val
	hashStoreMtx.Unlock()

	return parser.Value{Typ: parser.BULK_STRING, Bulk: "OK"}
}

func hget(args []parser.Value) parser.Value {
	if len(args) != 2 {
		return parser.Value{Typ: parser.SIMPLE_ERROR, Str: "ERR wrong number of arguments for 'hget' command"}
	}

	hash := args[0].Bulk
	key := args[1].Bulk

	hashStoreMtx.RLock()
	val, ok := hashStore[hash][key]
	hashStoreMtx.RUnlock()

	if !ok {
		return parser.Value{Typ: parser.NULL}
	}
	return parser.Value{Typ: parser.BULK_STRING, Bulk: val}
}

// TODO: implement hgetAll
// func hgetAll(args []parser.Value) parser.Value {}
