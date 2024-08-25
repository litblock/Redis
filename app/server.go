package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

//var Commands = map[string]func([]Value) Value{
//	"PING": ping
//	"ECHO": echo
//}

type Entry struct {
	Value       string
	TimeCreated time.Time
	ExpiresAt   time.Time
}

type Cache map[string]*Entry

var cache = Cache{}

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
			expiry := time.Time{}
			key := split[4]
			value := split[6]
			now := time.Now()
			log.Println(split[8])
			log.Println(split[10])
			duration, _ := strconv.ParseInt(split[10], 10, 64)
			switch split[8] {
			case "px":
				expiry = now.Add(time.Duration(duration) * time.Millisecond)
			case "ex":
				expiry = now.Add(time.Duration(duration) * time.Second)
			default:
				return "Time invalid"
			}
			cache[key] = &Entry{
				Value:       value,
				TimeCreated: now,
				ExpiresAt:   expiry,
			}
			return "+OK\r\n"
		} else if command == "GET" {
			key := split[4]
			value, found := cache[key]
			if !found || (value.ExpiresAt.Before(time.Now()) && value.ExpiresAt != time.Time{}) {
				return "$-1\r\n"
			}
			return fmt.Sprintf("$%d\r\n%v\r\n", len(value.Value), value.Value)
		}
	}
	return ""
}
