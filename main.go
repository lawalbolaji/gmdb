package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"

	"gmdb/commands"
	"gmdb/parser"
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
	const PORT = "6379"
	fmt.Println("Listening on Port:", PORT)

	// create tcp server
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", PORT))
	if err != nil {
		log.Fatal(err)
		return
	}

	// accept incoming connections
	conn, err := listener.Accept()
	if err != nil {
		log.Fatal(err)
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

		if ast.Typ != parser.ARRAY {
			fmt.Println("Invalid request, expected array")
			continue
		}

		if len(ast.Array) == 0 {
			fmt.Println("Invalid request, expected array length > 0")
			continue
		}

		command := strings.ToUpper(ast.Array[0].Bulk)
		args := ast.Array[1:]

		writer := parser.NewWriter(conn)

		handler, ok := commands.Handlers[command]
		if !ok {
			fmt.Println("Invalid command: ", command)
			writer.Write(parser.Value{Typ: parser.SIMPLE_STRING, Str: ""})
			continue
		}

		writer.Write(handler(args))
	}
}
