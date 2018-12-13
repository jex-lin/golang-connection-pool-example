package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"time"
)

const (
	CONN_HOST = "localhost"
	CONN_PORT = "3333"
	CONN_TYPE = "tcp"
)

func launchTCPserver() {
	// Listen for incoming connections.
	l, err := net.Listen(CONN_TYPE, CONN_HOST+":"+CONN_PORT)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	// Close the listener when the application closes.
	defer l.Close()
	fmt.Println("Listening on " + CONN_HOST + ":" + CONN_PORT)
	for {
		// Listen for an incoming connection.
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}
		fmt.Println("[server] connection established")

		// Handle connections in a new goroutine.
		go handleRequest(conn)
	}
}

// Handles incoming requests.
func handleRequest(conn net.Conn) {
OuterLoop:
	for {
		// will listen for message to process ending in newline (\n)
		message, err := bufio.NewReader(conn).ReadString('\n')
		switch err {
		case io.EOF:
			fmt.Printf("[server] client disconnected: %v\n", err)
			break OuterLoop
		case nil:
		default:
			fmt.Println(err)
			break OuterLoop
		}
		// output message received
		fmt.Print("[server] Message Received:", string(message))
		// Processing
		time.Sleep(1 * time.Second)
		// send new string back to client
		conn.Write([]byte(string(message) + "\n"))
	}
}
