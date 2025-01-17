package main

import (
	"fmt"
	"net"
	"os"

	"github.com/alphahorizonio/tinynet/pkg/tinynet"
)

var (
	LADDR  = "127.0.0.1:1234"
	BUFLEN = 1024
)

func main() {
	lis, err := tinynet.Listen("tcp", LADDR)
	if err != nil {
		fmt.Println("could not listen", err)

		os.Exit(1)
	}

	fmt.Println("Listening on", LADDR)

	for {
		conn, err := lis.Accept()
		if err != nil {
			fmt.Println("could not accept", err)

			os.Exit(1)
		}

		fmt.Println("Client connected")

		go func(innerConn net.Conn) {
			for {
				buf := make([]byte, BUFLEN)
				if n, err := innerConn.Read(buf); err != nil {
					if n == 0 {
						break
					}

					fmt.Println("could not read from connection, removing connection", err)

					break
				}

				out := []byte(fmt.Sprintf("You've sent: %v", string(buf)))
				if n, err := innerConn.Write(out); err != nil {
					if n == 0 {
						break
					}

					fmt.Println("could not write from connection, removing connection", err)

					break
				}
			}

			fmt.Println("Client disconnected")

			if err := innerConn.Close(); err != nil {
				fmt.Println("could not close connection", err)
			}

			return
		}(conn)
	}
}
