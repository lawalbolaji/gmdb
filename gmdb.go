package main

import (
	"github.com/lawalbolaji/gmdb/tcp"
)

/*
	(docs) https://redis.io/docs/reference/protocol-spec/#bulk-strings
	TODO:
	- implement pub/sub: subscribe to channels, send messages, etc.
		- implement rdb persistence
		- allow users to choose persistence modes
*/

func main() {
	const port = "6379"
	tcp.CreateConnectionManager(port)
}
