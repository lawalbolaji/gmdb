package tcp

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"github.com/lawalbolaji/gmdb/commands"
	"github.com/lawalbolaji/gmdb/parser"
	"github.com/lawalbolaji/gmdb/store"
	"github.com/lawalbolaji/gmdb/transaction"
)

func CreateConnectionManager(port string) {
	fmt.Println("Listening on Port:", port)

	// create tcp server
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatal(err)
	}

	/* accept incoming connections concurrently */
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}

		go func(conn net.Conn) {
			fmt.Println("\n---New connection----")
			defer conn.Close()
			handleIncomingMessages(conn)
		}(conn)
	}
}

func handleIncomingMessages(conn net.Conn) {
	isInTrxMode := false
	queue := transaction.NewCommandQueue()

	for {
		ast, respParsedSuccessfully, exit := readRESPToAst(conn)
		if !respParsedSuccessfully {
			if exit {
				return
			}
			// try again
			continue
		}

		bufWriter := bufio.NewWriter(conn)
		writer := parser.NewWriter(bufWriter) // buffer output to efficient io for transaction mode
		if isInTrxMode {
			command := strings.ToUpper(ast.Array[0].Bulk)
			stayInTrxMode, output := transaction.HandleCommandInTransactionMode(ast, queue, command)

			isInTrxMode = stayInTrxMode
			writer.Write(output)

			bufWriter.Flush()
			continue
		}

		command := strings.ToUpper(ast.Array[0].Bulk)
		args := ast.Array[1:]

		// if command is "multi", the handler simple acknowledges the request
		if command == "MULTI" {
			isInTrxMode = true
		}

		locks := store.GetRequiredLocks([][]parser.Value{ast.Array})
		handler, ok := commands.Handlers[command]
		if !ok {
			writer.Write(parser.Value{Typ: parser.SIMPLE_STRING, Str: "command not supported"})
			bufWriter.Flush()
			continue
		}

		// we do not need a lock for the matched operation
		if len(locks) == 0 {
			writer.Write(handler(args))
			bufWriter.Flush()
			continue
		}

		locks[0].Lock()
		writer.Write(handler(args))
		locks[0].Unlock()
		bufWriter.Flush()
	}
}

func readRESPToAst(conn net.Conn) (rsp parser.Value, ok bool, exit bool) {
	const readDeadline = 1 * time.Minute // if client doesn't send a request on an active connection for 1 minute, close the connection
	err := conn.SetReadDeadline(time.Now().Add(readDeadline))
	if err != nil {
		log.Println("unable to retrieve connection")
		return parser.Value{}, false, true
	}

	resp := parser.NewResp(conn)
	ast, err := resp.Read()
	if err != nil {
		if tErr, ok := err.(net.Error); ok && tErr.Timeout() {
			conn.Write([]byte("+timeout\r\n")) // closing the connection causes a broken pipe in the client, research timeout mechanism in redis to make this compatible with the client
		} else if err.Error() == "EOF" {
			log.Println("connection closed")
		} else {
			log.Println(err)
		}
		return parser.Value{}, false, true // close connection
	}

	if err := validateRESPPayload(ast, conn); err != nil {
		return parser.Value{}, false, false
	}

	return ast, true, false
}

func validateRESPPayload(ast parser.Value, conn net.Conn) error {
	if ast.Typ != parser.ARRAY || len(ast.Array) == 0 {
		conn.Write([]byte("+invalid format type\r\n"))
		return errors.New("invalid resp format")
	}

	return nil
}
