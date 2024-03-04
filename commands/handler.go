package commands

import (
	"gmdb/parser"
	"gmdb/store"
)

func ping(args []parser.Value) parser.Value {
	if len(args) == 0 {
		return parser.Value{Typ: parser.SIMPLE_STRING, Str: "PONG"}
	}

	return parser.Value{Typ: parser.BULK_STRING, Bulk: args[0].Bulk}
}

var Handlers = map[string]func([]parser.Value) parser.Value{
	"PING":    ping,
	"SET":     set,
	"GET":     get,
	"HGET":    hget,
	"HSET":    hset,
	"HGETALL": hgetAll,
	"MULTI":   multi,
}

func set(args []parser.Value) parser.Value {
	if len(args) != 2 {
		return parser.Value{Typ: parser.SIMPLE_ERROR, Str: "ERR wrong number of arguments for 'set' command"}
	}

	key := args[0].Bulk
	val := args[1].Bulk

	store.SetValInKVStore(key, val)
	return parser.Value{Typ: parser.SIMPLE_STRING, Str: "OK"}
}

func get(args []parser.Value) parser.Value {
	if len(args) != 1 {
		return parser.Value{Typ: parser.SIMPLE_ERROR, Str: "ERR wrong number of arguments for 'get' command"}
	}

	key := args[0].Bulk
	value, ok := store.GetValFromKVStore(key)

	if !ok {
		return parser.Value{Typ: parser.NULL}
	}
	return parser.Value{Typ: parser.BULK_STRING, Bulk: value}
}

func hset(args []parser.Value) parser.Value {
	if len(args) != 3 { // hashmap, key, val
		return parser.Value{Typ: parser.SIMPLE_ERROR, Str: "ERR wrong number of arguments for 'hset' command"}
	}

	key := args[0].Bulk
	hash := args[1].Bulk
	val := args[2].Bulk

	store.SetValInHashStore(key, hash, val)
	return parser.Value{Typ: parser.BULK_STRING, Bulk: "OK"}
}

func hget(args []parser.Value) parser.Value {
	if len(args) != 2 {
		return parser.Value{Typ: parser.SIMPLE_ERROR, Str: "ERR wrong number of arguments for 'hget' command"}
	}

	key := args[0].Bulk
	hash := args[1].Bulk

	val, ok := store.GetValFromHashStore(key, hash)
	if !ok {
		return parser.Value{Typ: parser.NULL}
	}

	return parser.Value{Typ: parser.BULK_STRING, Bulk: val}
}

func hgetAll(args []parser.Value) parser.Value {
	if len(args) != 1 {
		return parser.Value{Typ: parser.SIMPLE_ERROR, Str: "ERR wrong number of arguments for 'hgetall' command"}
	}

	key := args[0].Bulk
	keysAndVal, err := store.GetAllKeysAndValFromHashStore(key)
	if err != nil {
		return parser.Value{Typ: parser.BULK_STRING, Bulk: err.Error()}
	}

	result := parser.Value{Typ: parser.ARRAY}
	for _, val := range keysAndVal {
		result.Array = append(result.Array, parser.Value{Typ: parser.BULK_STRING, Bulk: val})
	}

	return result
}

func multi(args []parser.Value) parser.Value {
	if len(args) != 0 {
		return parser.Value{Typ: parser.SIMPLE_ERROR, Str: "ERR multi does not take arguments"}
	}

	return parser.Value{Typ: parser.BULK_STRING, Bulk: "ok"}
}
