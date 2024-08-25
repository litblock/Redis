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
		log.Println("Failed to bind to port 6379")
		os.Exit(1)
	}

	for {
		connection, err := l.Accept()
		if err != nil {
			log.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		go handleRequest(connection)
	}
}

func handleRequest(connection net.Conn) {
	buf := make([]byte, 1024)
	for {
		_, err := connection.Read(buf)
		if err != nil {
			log.Println("Error while trying to read:", err.Error())
		}
		msg := parseToString(buf)
		if msg == "" {
			log.Println("Invalid Command")
		}
		_, err = connection.Write([]byte(msg))
		if err != nil {
			log.Println("Error trying to write", err.Error())
		}
	}
}

func parseToString(buf []byte) string {
	input := (string)(buf)
	split := strings.Split(input, "\r\n")
	log.Println(split)
	if split[0][0] == '*' {
		command := strings.ToUpper(split[2])
		if command == "ECHO" {
			var out strings.Builder
			for i := 3; i < len(split)-1; i++ {
				out.WriteString(split[i] + "\r\n")
			}
			return out.String()
		} else if command == "PING" {
			return "+PONG\r\n"
		} else if command == "SET" {
			
		} else if command == "GET" {

		}
	}
	return ""
}
