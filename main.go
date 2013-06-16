package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
)

func main() {
	log.Printf("Server started\n")
	addr := ":8080"
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Printf("Error: listen(): %s", err)
		os.Exit(1)
	}

	log.Printf("Accepting connections at: %s", addr)
	store := &store{
		data: make(map[string]string),
		lock: &sync.RWMutex{},
	}

	var id int64
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error: Accept(): %s", err)
			continue
		}

		id++
		client := &client{id: id, conn: conn, store: store}
		go client.serve()
	}
}

type store struct {
	data map[string]string
	lock *sync.RWMutex
}

func (store *store) Get(key string) string {
	store.lock.RLock()
	defer store.lock.RUnlock()

	return store.data[key]
}

func (store *store) Set(key string, val string) {
	store.lock.Lock()
	defer store.lock.Unlock()

	store.data[key] = val
}

type client struct {
	id     int64
	conn   net.Conn
	reader *bufio.Reader
	store  *store
}

func (client *client) serve() {
	defer client.conn.Close()

	client.log("Accepted connection: %s", client.conn.LocalAddr())
	client.reader = bufio.NewReader(client.conn)

	for {
		cmd, err := client.readCommand()
		if err != nil {
			if err == io.EOF {
				client.log("Disconnected")
			} else if _, ok := err.(protocolError); ok {
				client.sendError(err)
			} else {
				client.logError("readCommand(): %s", err)
			}
			return
		}

		switch cmd.Name {
		case "GET":
			if len(cmd.Args) < 1 {
				client.sendError(fmt.Errorf("GET expects 1 argument"))
				return
			}
			val := client.store.Get(cmd.Args[0])
			client.send(val)
		case "SET":
			if len(cmd.Args) < 1 {
				client.sendError(fmt.Errorf("SET expects 2 arguments"))
				return
			}
			client.store.Set(cmd.Args[0], cmd.Args[1])
			fmt.Fprintf(client.conn, "+OK\r\n")
		default:
			client.sendError(fmt.Errorf("unkonwn command: %s", cmd.Name))
		}
	}
}

func (client *client) log(msg string, args ...interface{}) {
	prefix := fmt.Sprintf("Client #%d: ", client.id)
	log.Printf(prefix+msg, args...)
}

func (client *client) logError(msg string, args ...interface{}) {
	client.log("Error: "+msg, args...)
}

func (client *client) send(val string) {
	fmt.Fprintf(client.conn, "$%d\r\n%s\r\n", len(val), val)
}

func (client *client) sendError(err error) {
	client.logError(err.Error())
	client.sendLine("-ERR " + err.Error() + "\r\n")
}

func (client *client) sendLine(line string) {
	if _, err := io.WriteString(client.conn, line); err != nil {
		client.log("Error for client.sendLine(): %s", err)
	}
}

type protocolError string

func (e protocolError) Error() string {
	return string(e)
}

func (client *client) readCommand() (*command, error) {
	for {
		line, err := client.readLine()
		if err != nil {
			return nil, err
		}

		// Example: *5 (command consisting of 5 arguments)
		if !strings.HasPrefix(line, "*") {
			return &command{Name: line}, nil
		}

		argcStr := line[1:]
		argc, err := strconv.ParseUint(argcStr, 10, 64)
		if err != nil || argc <= 1 {
			return nil, protocolError("invalid argument count: " + argcStr)
		}

		args := make([]string, 0, argc)
		for i := 0; i < int(argc); i++ {
			line, err := client.readLine()
			if err != nil {
				return nil, err
			}

			// Example: $3 (next line has 3 bytes + \r\n)
			if !strings.HasPrefix(line, "$") {
				return nil, protocolError("unknown command: " + line)
			}

			argLenStr := line[1:]
			argLen, err := strconv.ParseUint(argLenStr, 10, 64)
			if err != nil {
				return nil, protocolError("invalid argument length: " + argLenStr)
			}

			arg := make([]byte, argLen+2)
			if _, err := io.ReadFull(client.reader, arg); err != nil {
				return nil, err
			}

			args = append(args, string(arg[0:len(arg)-2]))
		}

		return &command{Name: args[0], Args: args[1:]}, nil
	}
}

func (client *client) readLine() (string, error) {
	var line string
	for {
		partialLine, isPrefix, err := client.reader.ReadLine()
		if err != nil {
			return "", err
		}

		line += string(partialLine)
		if isPrefix {
			continue
		}

		return line, nil
	}
}

type command struct {
	Name string
	Args []string
}
