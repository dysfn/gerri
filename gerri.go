package main

/*
Minimal IRC bot in Go

TODO:
* add more plugins
* store connection info in json file
*/

import (
	"fmt"
	"log"
	"bufio"
	"net"
	"net/textproto"
	"strings"
	"net/url"
	"encoding/json"
	"net/http"
	"io/ioutil"
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

/* structs */
type Privmsg struct {
	Source string
	Target string
	Message []string
}

type DuckDuckGo struct {
	AbstractText string
	AbstractURL string
}

/* simple message builders */
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

/* plugin helpers */
func queryDuckDuckGo(term string) *DuckDuckGo {
	var ddg *DuckDuckGo = &DuckDuckGo{}

	encoded := url.QueryEscape(term)
	resource := fmt.Sprintf("http://api.duckduckgo.com?format=json&q=%s", encoded)

	resp, err := http.Get(resource)
	if err != nil {
		log.Fatal(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	if err = json.Unmarshal(body, ddg); err != nil {
		log.Fatal(err)
	}

	return ddg
}

/* plugins */
func replyPing(msg string) string {
	return "meow"
}

func replyWik(msg string) string {
	ddg := queryDuckDuckGo(msg)
	if ddg.AbstractText != "" && ddg.AbstractURL != "" {
		truncated := strings.Split(ddg.AbstractText, " ")[:50]  // first 50 words
		return fmt.Sprintf("%s... (source: %s)", strings.Join(truncated, " "), ddg.AbstractURL)
	} else {
		return "(no results found)"
	}
}

var repliers = map[string]func(string) string{
	":!ping": replyPing,
	":!wik": replyWik,
}

func buildPrivmsg(pm Privmsg) string {
	/* replies PRIVMSG message */
	msg := strings.Join(pm.Message[1:], " ")
	fn, found := repliers[pm.Message[0]]
	if found {
		return msgPrivmsg(pm.Target, fn(msg))
	} else {
		return ""
	}
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
				pm := Privmsg{Source: tokens[0], Target: tokens[2], Message: tokens[3:]}
				reply := buildPrivmsg(pm)
				if reply != "" {
					log.Printf("reply: %s", reply)
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