package main

/*
Minimal IRC bot in Go

TODO:
* reply PING :cookie (e.g. PING :sinisalo.freenode.net) with PING: cookie
* add plugins (!wik, !random, !ping, !title)
* store connection info in json file
*/

import (
        "fmt"
        "log"
	"net"
        "bufio"
        "net/textproto"
)

const (
	USER = "USER"
	NICK = "NICK"
	JOIN = "JOIN"
	SUFFIX = "\r\n"
)

func connect(server string, port string) (net.Conn, error) {
	/* establishes irc connection  */
	log.Printf("connecting to %s:%s...", server, port)
	conn, err := net.Dial("tcp", server + ":" + port)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("connected")
	return conn, err
}

func read(ch chan<- string, conn net.Conn) {
	/* defines goroutine sending lines to channel */
	reader := textproto.NewReader(bufio.NewReader(conn))
	for {
		line, err := reader.ReadLine()
		if err != nil {
			log.Fatal(err)
			break
		}
		ch <- line
	}
}

func write(ch <-chan string, conn net.Conn) {
	/* defines goroutine receiving lines from channel */
	for {
		line, ok := <-ch
		if !ok {
			log.Fatal("aborted: failed to receive from channel")
			break
		}
		fmt.Println(line)
	}
}

func main() {
        server, port := "chat.freenode.net", "8002"
	nick, channel := "gerri", "#microamp"

        // connect to irc
	conn, err := connect(server, port)
	if err != nil {
		log.Fatal(err)
	}

	// send messages: USER/NICK/JOIN
	conn.Write([]byte(USER + " " + nick + " 8 * :" + nick + SUFFIX))
	conn.Write([]byte(NICK + " " + nick + SUFFIX))
	conn.Write([]byte(JOIN + " " + channel + SUFFIX))

	defer conn.Close()

        // define goroutines communicating via channel
	ch := make(chan string)
	go read(ch, conn)
	go write(ch, conn)

	var input string
	fmt.Scanln(&input)
}
