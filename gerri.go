package main

/*
Minimal IRC bot in Go

TODO:
* google app engine integration to evaluate python code
* add more plugins (!title, !hn, !reddit, ...)
* store connection info in json file
* separate out plugins from main program
*/

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"net"
	"net/http"
	"net/textproto"
	"net/url"
	"strings"
	"time"
)

const (
	VERSION = "0.1.1"
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
	JIRA = "https://webdrive.atlassian.net"
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

	if term == "" {
		term = "cat"
	}
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

func timeDelta(weekday string, hour int, minute int) string {
	now := time.Now()
	wd := now.Weekday().String()
	if wd == weekday {
		y, m, d := now.Date()
		location := now.Location()

		beertime := time.Date(y, m, d, hour, minute, 0, 0, location)
		diff := beertime.Sub(now)

		if diff.Seconds() > 0 {
			return fmt.Sprintf("less than %d minute(s) to go...", int(math.Ceil(diff.Minutes())))
		}
		return "it's beertime!"
	}
	return fmt.Sprintf("it's only %s...", strings.ToLower(wd))
}

/* plugins */
func replyVer(pm Privmsg) string {
	return msgPrivmsg(pm.Target, fmt.Sprintf("gerri version: %s", VERSION))
}

func replyPing(pm Privmsg) string {
	return msgPrivmsg(pm.Target, "meow")
}

func replyGIF(pm Privmsg) string {
	msg := strings.Join(pm.Message[1:], " ")
	giphy := searchGiphy(msg)
	if len(giphy.Data) > 0 {
		m := fmt.Sprintf("http://media.giphy.com/media/%s/giphy.gif", giphy.Data[rand.Intn(len(giphy.Data))].ID)
		return msgPrivmsg(pm.Target, m)
	}
	return msgPrivmsg(pm.Target, "(zzzzz...)")
}

func replyDay(pm Privmsg) string {
	return msgPrivmsg(pm.Target, strings.ToLower(time.Now().Weekday().String()))
}

func replyWik(pm Privmsg) string {
	msg := strings.Join(pm.Message[1:], " ")
	if strings.TrimSpace(msg) != "" {
		ddg := queryDuckDuckGo(msg)
		if ddg.AbstractText != "" && ddg.AbstractURL != "" {
			size := 30
			words := strings.Split(ddg.AbstractText, " ")
			var m string
			if len(words) > size {
				m = fmt.Sprintf("%s... (source: %s)", strings.Join(words[:size], " "), ddg.AbstractURL)
			} else {
				m = fmt.Sprintf("%s (source: %s)", ddg.AbstractText, ddg.AbstractURL)
			}
			return msgPrivmsg(pm.Target, m)
		}
		return msgPrivmsg(pm.Target, "(zzzzz...)")
	}
	return ""
}

func replyBeertime(pm Privmsg) string {
	return msgPrivmsg(pm.Target, timeDelta(BEERTIME_WD, BEERTIME_HR, BEERTIME_MIN))
}

func replyJira(pm Privmsg) string {
	msg := strings.Join(pm.Message[1:], " ")
	if strings.TrimSpace(msg) != "" {
		return msgPrivmsg(pm.Target, JIRA + "/browse/" + strings.ToUpper(msg))
	}
	return msgPrivmsg(pm.Target, JIRA)
}

func replyAsk(pm Privmsg) string {
	msg := strings.Join(pm.Message[1:], " ")
	if strings.TrimSpace(msg) != "" {
		rand.Seed(time.Now().UnixNano())
		return msgPrivmsg(pm.Target, [2]string{"yes!", "no..."}[rand.Intn(2)])
	}
	return ""
}

var repliers = map[string]func(Privmsg) string {
	":!ver": replyVer,
	":!version": replyVer,
	":!ping": replyPing,
	":!day": replyDay,
	":!gif": replyGIF,
	":!wik": replyWik,
	":!beertime": replyBeertime,
	":!jira": replyJira,
	":!ask": replyAsk,
}

func buildReply(pm Privmsg) string {
	/* replies PRIVMSG message */
	fn, found := repliers[pm.Message[0]]
	if found {
		return fn(pm)
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
