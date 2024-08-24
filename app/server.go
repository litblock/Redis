package main

import (
	"fmt"
	"net"
	"os"
)

func main() {
	fmt.Println("Logs from your program will appear here!")

	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}
	connection, err := l.Accept()
	for {
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		handleRequest(connection)
	}
}

func handleRequest(connection net.Conn) {
	buf := make([]byte, 1024)
	_, err := connection.Read(buf)
	if err != nil {
		fmt.Println("Error while trying to read:", err.Error())
	}
	//fmt.Println(in)
	_, err = connection.Write([]byte("+PONG\r\n"))
	if err != nil {
		fmt.Println("Error trying to write", err.Error())
	}
}
