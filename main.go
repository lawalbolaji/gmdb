package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

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

	defer conn.Close()

	for {
		buf := make([]byte, 1024)

		// read msg from client
		_, err := conn.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}

			fmt.Println("error reading from client", err.Error())
			os.Exit(1)
		}

		fmt.Println(string(buf))
		conn.Write([]byte("+OK\r\n"))
	}
}
