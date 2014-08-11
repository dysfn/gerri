package main

/*
Minimal IRC bot in Go

TODO:
* add plugins (!wik, !random, !title)
* store connection info in json file
*/

import (
        "fmt"
        "log"
	"net"
        "bufio"
        "net/textproto"
	"strings"
)

const (
	USER = "USER"
	NICK = "NICK"
	JOIN = "JOIN"
	PING = "PING"
	PONG = "PONG"
	PRIVMSG = "PRIVMSG"
	SUFFIX = "\r\n"
)

type privmsg struct {
	src string
	tgt string
	msg []string
}

func msgUser(nick string) string {
	return USER + " " + nick + " 8 * :" + nick + SUFFIX
}

func msgNick(nick string) string {
	return NICK + " " + nick + SUFFIX
}

func msgJoin(channel string) string {
	return JOIN + " " + channel + SUFFIX
}

func msgPong(host string) string {
	return PONG + " :" + host + SUFFIX
}

func msgPrivmsg(receiver string, msg string) string {
	return PRIVMSG + " " + receiver + " :" + msg + SUFFIX
}

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

func handlePrivmsg(pm privmsg) string {
	msg := strings.Join(pm.msg, " ")
	if strings.HasPrefix(msg, ":!ping") {
		return msgPrivmsg(pm.tgt, "meow")
	}
	return ""
}

func send(ch chan<- string, conn net.Conn) {
	/* defines goroutine sending messages to channel */
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

func receive(ch <-chan string, conn net.Conn) {
	/* defines goroutine receiving messages from channel */
	for {
		line, ok := <-ch
		if !ok {
			log.Fatal("aborted: failed to receive from channel")
			break
		}
		log.Printf(line)

		if strings.HasPrefix(line, PING) {
			// reply PING with PONG
			msg := msgPong(strings.Split(line, ":")[1])
			conn.Write([]byte(msg))
			log.Printf(msg)
		} else {
			// reply PRIVMSG
			tokens := strings.Split(line, " ")
			if len(tokens) >= 4 && tokens[1] == PRIVMSG {
				pm := privmsg{src: tokens[0], tgt: tokens[2], msg: tokens[3:]}
				reply := handlePrivmsg(pm)
				log.Printf("reply: %s", reply)
				if reply != "" {
					conn.Write([]byte(reply))
				}
			}
		}
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
	conn.Write([]byte(msgUser(nick)))
	conn.Write([]byte(msgNick(nick)))
	conn.Write([]byte(msgJoin(channel)))

	defer conn.Close()

        // define goroutines communicating via channel
	ch := make(chan string)
	go send(ch, conn)
	go receive(ch, conn)

	var input string
	fmt.Scanln(&input)
}
