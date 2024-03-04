package main

import (
	"gmdb/tcp"
)

/*
	(docs) https://redis.io/docs/reference/protocol-spec/#bulk-strings
	TODO:
	- implement transactions
		- implement pub/sub: subscribe to channels, send messages, etc.
		- implement rdb persistence
		- allow users to choose persistence modes
*/

func main() {
	tcp.CreateConnectionManager("6379")
}
