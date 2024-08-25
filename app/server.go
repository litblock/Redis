package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

//var Commands = map[string]func([]Value) Value{
//	"PING": ping
//	"ECHO": echo
//}

func main() {
	fmt.Println("Logs from your program will appear here!")

	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}

	for {
		connection, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		go handleRequest(connection)
	}
}

func handleRequest(connection net.Conn) {
	buf := make([]byte, 1024)
	for {
		n, err := connection.Read(buf)
		if err != nil {
			fmt.Println("Error while trying to read:", err.Error())
		}
		log.Println("read: ", n)
		parseToString(buf)
		//_, err = connection.Write([]byte("+PONG\r\n"))
		if err != nil {
			fmt.Println("Error trying to write", err.Error())
		}
	}
}

func parseToString(buf []byte) {
	input := (string)(buf)
	split := strings.Split(input, "\r\n")
	log.Println(split)
}
