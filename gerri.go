package main

/*
Minimal IRC bot in Go
*/

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/dysfn/gerri/cmd"
	"github.com/dysfn/gerri/data"
	"github.com/dysfn/gerri/plugin"
	"io/ioutil"
	"log"
	"net"
	"net/textproto"
	"strings"
)

const (
	CONFIG = "config.json" // config filename
)

/* plugin mappings */
var repliers = map[string]func(data.Privmsg, *data.Config) (string, error){
	":!advice":   plugin.ReplyAdvice,
	":!ask":      plugin.ReplyAsk,
	":!beertime": plugin.ReplyBeertime,
	":!day":      plugin.ReplyDay,
	":!ddg":      plugin.ReplyDdg,
	":!gif":      plugin.ReplyGIF,
	":!jira":     plugin.ReplyJira,
	":!ping":     plugin.ReplyPing,
	":!py":       plugin.ReplyPy,
	":!quote":    plugin.ReplyQuote,
	":!slap":     plugin.ReplySlap,
	":!title":    plugin.ReplyTitle,
	":!time":     plugin.ReplyTime,
	":!ud":       plugin.ReplyUd,
	":!ver":      plugin.ReplyVer,
	":!version":  plugin.ReplyVer,
	":!wik":      plugin.ReplyWik,
	":!wa":       plugin.ReplyWA,
}

func buildReply(conn net.Conn, pm data.Privmsg) {
	/* replies PRIVMSG message */
	fn, found := repliers[pm.Message[0]]
	if found {
		reply, err := fn(pm, readConfig(CONFIG))
		if err != nil {
			log.Printf("error: %s", err)
		} else {
			if reply != "" {
				log.Printf("reply: %s", reply)
				conn.Write([]byte(reply))
			}
		}
	}
}

func connect(server string, port string) (net.Conn, error) {
	/* establishes irc connection  */
	log.Printf("connecting to %s:%s...", server, port)
	conn, err := net.Dial("tcp", server+":"+port)
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

		tokens := strings.Split(line, " ")
		if tokens[0] == cmd.PING {
			// reply PING with PONG
			msg := cmd.Pong(strings.Split(line, ":")[1])
			conn.Write([]byte(msg))
			log.Printf(msg)
		} else {
			// reply PRIVMSG
			if len(tokens) >= 4 && tokens[1] == cmd.PRIVMSG {
				pm := data.Privmsg{Source: tokens[0], Target: tokens[2], Message: tokens[3:]}
				go buildReply(conn, pm) // reply asynchronously
			}
		}
	}
}

func readConfig(filename string) *data.Config {
	/* reads config from file */
	file, e := ioutil.ReadFile(filename)
	if e != nil {
		fmt.Printf("File error: %v\n", e)
	}

	var config *data.Config = &data.Config{}
	if err := json.Unmarshal(file, config); err != nil {
		log.Fatal(err)
	}
	return config
}

func main() {
	// read config from file
	config := readConfig(CONFIG)

	// connect to quote db
	if len(config.QuoteDB) != 0 {
		plugin.QuoteDB = plugin.ConnectQuoteDB(config.QuoteDB)
		defer plugin.QuoteDB.Close()
	}

	// connect to irc
	conn, err := connect(config.Server, config.Port)
	if err != nil {
		log.Fatal(err)
	}

	// send messages: USER/NICK/JOIN
	conn.Write([]byte(cmd.User(config.Nick)))
	conn.Write([]byte(cmd.Nick(config.Nick)))
	conn.Write([]byte(cmd.Join(config.Channel)))

	defer conn.Close()

	// define goroutines communicating via channel
	ch := make(chan string)
	go send(ch, conn)
	go receive(ch, conn)

	var input string
	fmt.Scanln(&input)
}
