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
	"time"
)

const (
	USER = "USER"
	NICK = "NICK"
	JOIN = "JOIN"
	PING = "PING"
	PONG = "PONG"
	PRIVMSG = "PRIVMSG"
	SUFFIX = "\r\n"
	BEERTIME_WD = "Friday"
	BEERTIME_HR = 16
	BEERTIME_MIN = 30
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

type GIF struct {
	ID string
}

type Giphy struct {
	Data []GIF
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
func searchGiphy(term string) *Giphy{
	var giphy *Giphy = &Giphy{}

	encoded := url.QueryEscape(term)
	resource := fmt.Sprintf("http://api.giphy.com/v1/gifs/search?api_key=dc6zaTOxFJmzC&q=%s", encoded)

	resp, err := http.Get(resource)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	if err = json.Unmarshal(body, giphy); err != nil {
		log.Fatal(err)
	}
	return giphy
}

func queryDuckDuckGo(term string) *DuckDuckGo {
	var ddg *DuckDuckGo = &DuckDuckGo{}

	encoded := url.QueryEscape(term)
	resource := fmt.Sprintf("http://api.duckduckgo.com?format=json&q=%s", encoded)

	resp, err := http.Get(resource)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	if err = json.Unmarshal(body, ddg); err != nil {
		log.Fatal(err)
	}

	return ddg
}

func diff(weekday string, hour int, minute int) string {
	now := time.Now()
	wd := now.Weekday().String()
	if wd == weekday {
		y, m, d := now.Date()
		location := now.Location()

		beertime := time.Date(y, m, d, hour, minute, 0, 0, location)
		diff := beertime.Sub(now)

		if diff.Seconds() > 0 {
			return fmt.Sprintf("%d minutes to go...", int(diff.Minutes()))
		}
		return "It's beertime!"
	}
	return fmt.Sprintf("It's only %s...", wd)
}

/* plugins */
func replyPing(msg string) string {
	return "meow"
}

func replyDay(msg string) string {
	return time.Now().Weekday().String()
}

func replyGIF(msg string) string {
	giphy := searchGiphy(msg)
	if giphy.Data[0].ID != "" {
		return fmt.Sprintf("http://media.giphy.com/media/%s/giphy.gif", giphy.Data[0].ID)
	}
	return "(zzzzz...)"
}

func replyWik(msg string) string {
	ddg := queryDuckDuckGo(msg)
	if ddg.AbstractText != "" && ddg.AbstractURL != "" {
		size := 35
		words := strings.Split(ddg.AbstractText, " ")
		if len(words) > size {
			return fmt.Sprintf("%s... (source: %s)", strings.Join(words[:size], " "), ddg.AbstractURL)
		} else {
			return fmt.Sprintf("%s (source: %s)", ddg.AbstractText, ddg.AbstractURL)
		}
	}
	return "(zzzzz...)"
}

func replyBeertime(msg string) string {
	return diff(BEERTIME_WD, BEERTIME_HR, BEERTIME_MIN)
}

var repliers = map[string]func(string) string{
	":!ping": replyPing,
	":!day": replyDay,
	":!gif": replyGIF,
	":!wik": replyWik,
	":!beertime": replyBeertime,
}

func buildReply(pm Privmsg) string {
	/* replies PRIVMSG message */
	msg := strings.Join(pm.Message[1:], " ")
	fn, found := repliers[pm.Message[0]]
	if found {
		return msgPrivmsg(pm.Target, fn(msg))
	}
	return ""
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
				reply := buildReply(pm)
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
