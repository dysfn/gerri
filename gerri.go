package main

/*
Minimal IRC bot in Go
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
	VERSION = "0.2.1"
	USER = "USER"
	NICK = "NICK"
	JOIN = "JOIN"
	PING = "PING"
	PONG = "PONG"
	PRIVMSG = "PRIVMSG"
	ACTION = "ACTION"
	SUFFIX = "\r\n"
	BEERTIME_WD = "Friday"
	BEERTIME_HR = 16
	BEERTIME_MIN = 30
	WIK_WORDS = 25
	JIRA = "https://webdrive.atlassian.net"
	GIPHY = "http://media.giphy.com"
	GIPHY_API = "http://api.giphy.com"
	GIPHY_KEY = "dc6zaTOxFJmzC"
	DDG_API = "http://api.duckduckgo.com"
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

func msgPrivmsgAction(receiver string, msg string) string {
	return fmt.Sprintf("%s %s :\001%s %s\001%s", PRIVMSG, receiver, ACTION, msg, SUFFIX)
}

/* plugin helpers */
func searchGiphy(term string) (*Giphy, error) {
	var giphy *Giphy = &Giphy{}

	if term == "" {
		term = "cat"
	}
	encoded := url.QueryEscape(term)
	resource := fmt.Sprintf("%s/v1/gifs/search?api_key=%s&q=%s", GIPHY_API, GIPHY_KEY, encoded)

	resp, err := http.Get(resource)
	if err != nil {
		return giphy, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return giphy, err
	}
	if err = json.Unmarshal(body, giphy); err != nil {
		return giphy, err
	}
	return giphy, nil
}

func queryDuckDuckGo(term string) (*DuckDuckGo, error) {
	var ddg *DuckDuckGo = &DuckDuckGo{}

	encoded := url.QueryEscape(term)
	resource := fmt.Sprintf("%s?format=json&q=%s", DDG_API, encoded)

	resp, err := http.Get(resource)
	if err != nil {
		return ddg, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ddg, err
	}
	if err = json.Unmarshal(body, ddg); err != nil {
		return ddg, err
	}
	return ddg, nil
}

func timeDelta(weekday string, hour int, minute int) (string, error) {
	now := time.Now()
	wd := now.Weekday().String()
	if wd == weekday {
		y, m, d := now.Date()
		location := now.Location()

		beertime := time.Date(y, m, d, hour, minute, 0, 0, location)
		diff := beertime.Sub(now)

		if diff.Seconds() > 0 {
			return fmt.Sprintf("less than %d minute(s) to go...", int(math.Ceil(diff.Minutes()))), nil
		}
		return "it's beertime!", nil
	}
	return fmt.Sprintf("it's only %s...", strings.ToLower(wd)), nil
}

func slapAction(target string) (string, error) {
	actions := []string {
		"slaps", "kicks", "destroys", "annihilates", "punches",
		"roundhouse kicks", "rusty hooks", "pwns", "owns"}
	if strings.TrimSpace(target) != "" {
		selected_action := actions[rand.Intn(len(actions))]
		return fmt.Sprintf("%s %s", selected_action, target), nil
	}
	return "zzzzz...", nil
}

/* plugins */
func replyVer(pm Privmsg) (string, error) {
	return msgPrivmsg(pm.Target, fmt.Sprintf("gerri version: %s", VERSION)), nil
}

func replyPing(pm Privmsg) (string, error) {
	return msgPrivmsgAction(pm.Target, "meows"), nil
}

func replyGIF(pm Privmsg) (string, error) {
	msg := strings.Join(pm.Message[1:], " ")
	giphy, err := searchGiphy(msg)
	if err != nil {
		return "", err
	}
	if len(giphy.Data) > 0 {
		m := fmt.Sprintf("%s/media/%s/giphy.gif", GIPHY, giphy.Data[rand.Intn(len(giphy.Data))].ID)
		return msgPrivmsg(pm.Target, m), nil
	}
	return msgPrivmsgAction(pm.Target, "zzzzz..."), nil
}

func replyDay(pm Privmsg) (string, error) {
	return msgPrivmsg(pm.Target, strings.ToLower(time.Now().Weekday().String())), nil
}

func replyWik(pm Privmsg) (string, error) {
	msg := strings.Join(pm.Message[1:], " ")
	if strings.TrimSpace(msg) != "" {
		ddg, err := queryDuckDuckGo(msg)
		if err != nil {
			return "", err
		}
		if ddg.AbstractText != "" && ddg.AbstractURL != "" {
			words := strings.Split(ddg.AbstractText, " ")
			var m string
			if len(words) > WIK_WORDS {
				text := strings.Join(words[:WIK_WORDS], " ")
				m = fmt.Sprintf("%s... (source: %s)", text, ddg.AbstractURL)
			} else {
				m = fmt.Sprintf("%s (source: %s)", ddg.AbstractText, ddg.AbstractURL)
			}
			return msgPrivmsg(pm.Target, m), nil
		}
		return msgPrivmsgAction(pm.Target, "zzzzz..."), nil
	}
	return "", nil
}

func replyBeertime(pm Privmsg) (string, error) {
	td, err := timeDelta(BEERTIME_WD, BEERTIME_HR, BEERTIME_MIN)
	if err != nil {
		return "", err
	}
	return msgPrivmsg(pm.Target, td), nil
}

func replyJira(pm Privmsg) (string, error) {
	msg := strings.Join(pm.Message[1:], " ")
	if strings.TrimSpace(msg) != "" {
		return msgPrivmsg(pm.Target, JIRA + "/browse/" + strings.ToUpper(msg)), nil
	}
	return msgPrivmsg(pm.Target, JIRA), nil
}

func replyAsk(pm Privmsg) (string, error) {
	msg := strings.Join(pm.Message[1:], " ")
	if strings.TrimSpace(msg) != "" {
		rand.Seed(time.Now().UnixNano())
		return msgPrivmsg(pm.Target, [2]string{"yes!", "no..."}[rand.Intn(2)]), nil
	}
	return "", nil
}

func replySlap(pm Privmsg) (string, error) {
	slap, err := slapAction(strings.Join(pm.Message[1:], " "))
	if err != nil {
		return "", err
	}
	return msgPrivmsgAction(pm.Target, slap), nil
}

var repliers = map[string]func(Privmsg) (string, error) {
	":!ver": replyVer,
	":!version": replyVer,
	":!ping": replyPing,
	":!day": replyDay,
	":!gif": replyGIF,
	":!wik": replyWik,
	":!beertime": replyBeertime,
	":!jira": replyJira,
	":!ask": replyAsk,
	":!slap": replySlap,
}

func buildReply(conn net.Conn, pm Privmsg) {
	/* replies PRIVMSG message */
	fn, found := repliers[pm.Message[0]]
	if found {
		reply, err := fn(pm)
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

		tokens := strings.Split(line, " ")
		if tokens[0] == PING {
			// reply PING with PONG
			msg := msgPong(strings.Split(line, ":")[1])
			conn.Write([]byte(msg))
			log.Printf(msg)
		} else {
			// reply PRIVMSG
			if len(tokens) >= 4 && tokens[1] == PRIVMSG {
				pm := Privmsg{Source: tokens[0], Target: tokens[2], Message: tokens[3:]}
				go buildReply(conn, pm)  // reply asynchronously
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
