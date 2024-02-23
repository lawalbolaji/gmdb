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
}

var setsHashmap = map[string]string{}
var setsMtx = sync.RWMutex{}

func set(args []parser.Value) parser.Value {
	if len(args) != 2 {
		return parser.Value{Typ: parser.SIMPLE_ERROR, Str: "ERR wrong number of arguments for 'set' command"}
	}

	key := args[0].Bulk
	val := args[1].Bulk

	setsMtx.Lock()
	setsHashmap[key] = val
	setsMtx.Unlock()

	return parser.Value{Typ: parser.SIMPLE_STRING, Str: "OK"}
}

func get(args []parser.Value) parser.Value {
	if len(args) != 1 {
		return parser.Value{Typ: parser.SIMPLE_ERROR, Str: "ERR wrong number of arguments for 'get' command"}
	}

	key := args[0].Bulk

	setsMtx.RLock()
	value, ok := setsHashmap[key]
	setsMtx.RUnlock()

	if !ok {
		return parser.Value{Typ: parser.NULL}
	}
	return parser.Value{Typ: parser.BULK_STRING, Bulk: value}
}
