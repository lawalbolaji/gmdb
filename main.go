package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"

	"gmdb/parser"

	"github.com/spewerspew/spew"
)

/*
	(docs) https://redis.io/docs/reference/protocol-spec/#bulk-strings
	TODO:
		- add support for pipelines
		- add batching
		- implement pub/sub: subscribe to channels, send messages, etc.
		- implement transactions
*/

func main() {
	const PORT = "6379"
	fmt.Println("Listening on Port:", PORT)

	// create tcp server
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", PORT))
	if err != nil {
		log.Fatal(err)
		return
	}

	// accept incoming connections
	conn, conErr := listener.Accept()
	if conErr != nil {
		log.Fatal(conErr)
		return
	}

	fmt.Println("\n---New connection----")
	defer conn.Close()

	for {
		// read msg from client
		resp := parser.NewResp(conn)
		ast, err := resp.Read()
		if err != nil {
			if err == io.EOF {
				break
			}

			fmt.Println("error reading from client", err.Error())
			os.Exit(1)
		}

		spew.Dump(ast)

		// just ack request for now
		parser.Respond(conn)
	}
}
