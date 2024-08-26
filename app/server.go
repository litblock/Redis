package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

//took from examples
const (
	opCodeModuleAux    byte = 247 /* Module auxiliary data. */
	opCodeIdle         byte = 248 /* LRU idle time. */
	opCodeFreq         byte = 249 /* LFU frequency. */
	opCodeAux          byte = 250 /* RDB aux field. */
	opCodeResizeDB     byte = 251 /* Hash table resize hint. */
	opCodeExpireTimeMs byte = 252 /* Expire time in milliseconds. */
	opCodeExpireTime   byte = 253 /* Old expire time in seconds. */
	opCodeSelectDB     byte = 254 /* DB number of the following keys. */
	opCodeEOF          byte = 255
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

var dir = flag.String("dir", "", "Directory to store RDB file")
var dbFileName = flag.String("dbfilename", "dump.rdb", "RDB file name")

func main() {
	flag.Parse()
	fmt.Println("Logs from your program will appear here!")

	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		log.Println("Failed to bind to port 6379")
		os.Exit(1)
	}
	loadRDB()
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
	//log.Println(split)
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
			log.Println(len(split))
			if len(split) > 8 {
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
		} else if command == "CONFIG" {
			//log.Println(split)
			log.Println(*dir)
			if split[4] == "GET" && split[6] == "dir" {
				return fmt.Sprintf("*2\r\n$3\r\ndir\r\n$%d\r\n%s\r\n", len(*dir), *dir)
			} else if split[4] == "GET" && split[6] == "dbfilename" {
				return fmt.Sprintf("*2\r\n$10\r\ndbfilename\r\n$%d\r\n%s\r\n", len(*dbFileName), *dbFileName)
			} else {
				return "Error"
			}
		} else if command == "KEYS" {
			//log.Println(cache)
			if split[4] == "*" {
				
				out := "*2\r\n"
				for key, value := range cache {
					log.Println(key, value)

				}
				return out
			} else {
				key := split[4]
				value, found := cache[key]
				if !found || (value.ExpiresAt.Before(time.Now()) && value.ExpiresAt != time.Time{}) {
					return "$-1\r\n"
				}
				return fmt.Sprintf("$%d\r\n%v\r\n", len(value.Value), value.Value)
			}
		}
	}
	return ""
}

func sliceIndex(data []byte, sep byte) int {
	for i, b := range data {
		if b == sep {
			return i
		}
	}
	return -1
}
func parseTable(bytes []byte) []byte {
	start := sliceIndex(bytes, opCodeResizeDB)
	end := sliceIndex(bytes, opCodeEOF)
	return bytes[start+1 : end]
}

func loadRDB() {
	log.Println("loading rdb")
	data, err := os.ReadFile(*dir + "/" + *dbFileName)
	if err != nil {
		log.Println("error opening file")
	}
	if len(data) == 0 {
		return
	}
	line := parseTable(data)
	key := line[4 : 4+line[3]]
	value := line[5+line[3]:]

	cache[string(key)] = &Entry{
		Value:       string(value),
		TimeCreated: time.Now(),
		ExpiresAt:   time.Time{},
	}
}
