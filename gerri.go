package main

/*
Minimal IRC bot in Go

TODO:
* store connection info in json file
* set up goroutines for reading/writing buffers
* declare const(ant) values for commands (e.g. NICK)
* reply PING :cookie (e.g. PING :sinisalo.freenode.net) with PING: cookie
* add plugins (!wik, !random, !ping, !title)
*/

import (
        "fmt"
        "log"
	"net"
        "bufio"
        "net/textproto"
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

func main(){
        server, port := "chat.freenode.net", "8002"
	nick, channel := "gerri", "#microamp"
	msgSuffix := "\r\n"

        // connect to irc
	conn, err := connect(server, port)
	if err != nil {
		log.Fatal(err)
	}

	// send messages; USER/NICK/JOIN
	conn.Write([]byte("USER " + nick + " 8 * :" + nick + msgSuffix))
	conn.Write([]byte("NICK " + nick + msgSuffix))
	conn.Write([]byte("JOIN " + channel + msgSuffix))

	defer conn.Close()

        // read buffers
	r := bufio.NewReader(conn)
	tp := textproto.NewReader(r)
	for {
		line, err := tp.ReadLine()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s%s", line, msgSuffix)
	}
}
