package tcp

import (
	"fmt"
	"gmdb/commands"
	"gmdb/parser"
	"log"
	"net"
	"strings"
	"time"
)

func CreateConnectionManager(port string) {
	fmt.Println("Listening on Port:", port)

	// create tcp server
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatal(err)
	}

	for {
		// accept incoming connections
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}

		go func(conn net.Conn) {
			fmt.Println("\n---New connection----")
			defer conn.Close()

			for {
				err := conn.SetReadDeadline(time.Now().Add(1 * time.Minute)) // underlying net.conn will error if the read happens after the timer has elapsed
				if err != nil {
					log.Println("unable to retrieve connection")
					return
				}

				resp := parser.NewResp(conn)
				ast, err := resp.Read()
				if err != nil {
					if tErr, ok := err.(net.Error); ok && tErr.Timeout() {
						conn.Write([]byte("+timeout\r\n")) // closing the connection cause a broken pipe in the client, research timeout mechanism in redis to make this compatible with the client
					} else if err.Error() == "EOF" {
						log.Println("connection closed")
					} else {
						log.Println(err)
					}
					return // close connection
				}

				// command mode
				if ast.Typ != parser.ARRAY {
					conn.Write([]byte("+wrong format type\r\n"))
					continue
				}

				if len(ast.Array) == 0 {
					conn.Write([]byte("+wrong format type\r\n"))
					continue
				}

				command := strings.ToUpper(ast.Array[0].Bulk)
				args := ast.Array[1:]

				writer := parser.NewWriter(conn)

				handler, ok := commands.Handlers[command]
				if !ok {
					writer.Write(parser.Value{Typ: parser.SIMPLE_STRING, Str: "command not supported"})
					continue
				}

				writer.Write(handler(args))
			}
		}(conn)
	}

}
