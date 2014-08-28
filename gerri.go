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
	"strconv"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

const (
	VERSION = "0.2.4"
	CONFIG = "config.json"	// config filename
	USER = "USER"
	NICK = "NICK"
	JOIN = "JOIN"
	PING = "PING"
	PONG = "PONG"
	PRIVMSG = "PRIVMSG"
	ACTION = "ACTION"
	SUFFIX = "\r\n"
)

var quoteDB *sql.DB = nil

/* structs */
type Config struct {
	Server string
	Port string
	Nick string
	Channel string
	WikMaxWords int
	Giphy string
	GiphyApi string
	DdgApi string
	GiphyKey string
	Jira string
	Beertime Beertime
	QuoteDB string
}

type Beertime struct {
	Day string
	Hour int
	Minute int
}

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
func searchGiphy(term string, config *Config) (*Giphy, error) {
	var giphy *Giphy = &Giphy{}

	if term == "" {
		term = "cat"
	}
	encoded := url.QueryEscape(term)
	resource := fmt.Sprintf("%s/v1/gifs/search?api_key=%s&q=%s", config.GiphyApi, config.GiphyKey, encoded)

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

func queryDuckDuckGo(term string, config *Config) (*DuckDuckGo, error) {
	var ddg *DuckDuckGo = &DuckDuckGo{}

	encoded := url.QueryEscape(term)
	resource := fmt.Sprintf("%s?format=json&q=%s", config.DdgApi, encoded)

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
func replyVer(pm Privmsg, config *Config) (string, error) {
	return msgPrivmsg(pm.Target, fmt.Sprintf("gerri version: %s", VERSION)), nil
}

func replyPing(pm Privmsg, config *Config) (string, error) {
	return msgPrivmsgAction(pm.Target, "meows"), nil
}

func replyGIF(pm Privmsg, config *Config) (string, error) {
	msg := strings.Join(pm.Message[1:], " ")
	giphy, err := searchGiphy(msg, config)
	if err != nil {
		return "", err
	}
	if len(giphy.Data) > 0 {
		m := fmt.Sprintf("%s/media/%s/giphy.gif", config.Giphy, giphy.Data[rand.Intn(len(giphy.Data))].ID)
		return msgPrivmsg(pm.Target, m), nil
	}
	return msgPrivmsgAction(pm.Target, "zzzzz..."), nil
}

func replyDay(pm Privmsg, config *Config) (string, error) {
	return msgPrivmsg(pm.Target, strings.ToLower(time.Now().Weekday().String())), nil
}

func replyWik(pm Privmsg, config *Config) (string, error) {
	msg := strings.Join(pm.Message[1:], " ")
	if strings.TrimSpace(msg) != "" {
		ddg, err := queryDuckDuckGo(msg, config)
		if err != nil {
			return "", err
		}
		if ddg.AbstractText != "" && ddg.AbstractURL != "" {
			words := strings.Split(ddg.AbstractText, " ")
			var m string
			if len(words) > config.WikMaxWords {
				text := strings.Join(words[:config.WikMaxWords], " ")
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

func replyBeertime(pm Privmsg, config *Config) (string, error) {
	td, err := timeDelta(config.Beertime.Day, config.Beertime.Hour, config.Beertime.Minute)
	if err != nil {
		return "", err
	}
	return msgPrivmsg(pm.Target, td), nil
}

func replyJira(pm Privmsg, config *Config) (string, error) {
	msg := strings.Join(pm.Message[1:], " ")
	if strings.TrimSpace(msg) != "" {
		return msgPrivmsg(pm.Target, config.Jira + "/browse/" + strings.ToUpper(msg)), nil
	}
	return msgPrivmsg(pm.Target, config.Jira), nil
}

func replyAsk(pm Privmsg, config *Config) (string, error) {
	msg := strings.Join(pm.Message[1:], " ")
	if strings.TrimSpace(msg) != "" {
		rand.Seed(time.Now().UnixNano())
		ors := strings.Split(msg, " or ")
		if len(ors) > 1 {
			return msgPrivmsg(pm.Target, ors[rand.Intn(len(ors))]), nil
		}
		return msgPrivmsg(pm.Target, [2]string{"yes!", "no..."}[rand.Intn(2)]), nil
	}
	return "", nil
}

func replySlap(pm Privmsg, config *Config) (string, error) {
	slap, err := slapAction(strings.Join(pm.Message[1:], " "))
	if err != nil {
		return "", err
	}
	return msgPrivmsgAction(pm.Target, slap), nil
}

func replyQuote(pm Privmsg, config *Config) (string, error) {
	// Establish if we want to just return a single quote
	if len(pm.Message) == 2 {
		quoteID, err := strconv.ParseInt(pm.Message[1], 0, 0)
		if err == nil && quoteID > 0 {
			quoteStr, err := getQuote(quoteID)
			if err != nil {
				return "", err
			}
			return msgPrivmsg(pm.Target, quoteStr), nil
		}
	}

	// Determine if this is a request for an existing quote, or attempting to
	// store a new one.
	msg := strings.Trim(strings.Join(pm.Message[1:], " "), " ")
	if strings.Count(msg, "\"") == 2 {
		if storeQuote(msg, pm.Source) {
			return msgPrivmsg(pm.Target, "Got it!"), nil
		}
		return "", nil
	}

	// At this point the assumption is that we just want a random quote
	quoteStr, err := getRandomQuote()
	if err != nil {
		return "", err
	}
	return msgPrivmsg(pm.Target, quoteStr), nil
}

func storeQuote(quoteStr string, sender string) (bool) {
	stmt, err := quoteDB.Prepare(`
		insert into quote(text, created_by, created_timestamp)
		values(?, ?, datetime('now'));
	`)
	if err != nil {
		log.Printf("Could not store quote. Failed to prepare statement: ", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(quoteStr, sender)
	if err != nil {
		log.Printf("Could not store quote: ", err)
		return false
	}

	return true
}

func connectQuoteDB(filename string) (*sql.DB) {
	// Opens and returns a database connection for the specificed sqlite3 DB.
	// If a DB does not already exist, it will be created.
	db, e := sql.Open("sqlite3", filename)
	if e != nil {
		log.Fatal(e)
	}

	if !checkQuoteDB(db) {
		setupQuoteDB(db)
	}
	return db
}

func checkQuoteDB(db *sql.DB) (bool) {
	// Check to see if this is a newly created DB or not by checking for the
	// existence of our main quote table.
	rows, e := db.Query(`
		select name from sqlite_master
		where type='table' and name='quote';
	`)
	if e != nil {
		log.Fatal("Unable to determine QuoteDB schema status: ", e)
	}
	defer rows.Close()
	return rows.Next()
}

func setupQuoteDB(db *sql.DB) {
	// Initialise the DB schema
	_, e := db.Exec(`
	create table quote
		(
			id integer not null primary key,
			text text,
			created_by text,
			created_timestamp text
		);
	`)

	if e != nil {
		log.Fatal("Unable to create QuoteDB schema: ", e)
	}
}

func getRandomQuote() (string, error) {
	row, err := quoteDB.Query(`select text from quote order by random() limit 1;`)
	if err != nil || !row.Next() {
		return "", err
	}
	defer row.Close()

	var quoteStr string
	row.Scan(&quoteStr)
	return quoteStr, err
}

func getQuote(quoteID int64) (string, error) {
	stmt, err := quoteDB.Prepare("select text from quote where id = (?);")
	if err != nil {
		log.Printf("Could not get quote. Failed to prepare statement: ", err)
	}
	defer stmt.Close()
	var quoteStr string
	err = stmt.QueryRow(quoteID).Scan(&quoteStr)
	if err != nil {
		return "", err
	}

	return quoteStr, err
}

var repliers = map[string]func(Privmsg, *Config) (string, error) {
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
	":!quote": replyQuote,
}

func buildReply(conn net.Conn, pm Privmsg) {
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

func readConfig(filename string) *Config {
	/* reads config from file */
	file, e := ioutil.ReadFile(filename)
	if e != nil {
		fmt.Printf("File error: %v\n", e)
	}

	var config *Config = &Config{}
	if err := json.Unmarshal(file, config); err != nil {
		log.Fatal(err)
	}
	return config
}

func main() {
	// read config from file
	config := readConfig(CONFIG)

	if len(config.QuoteDB) != 0 {
		quoteDB = connectQuoteDB(config.QuoteDB)
		defer quoteDB.Close()
	}

	// connect to irc
	conn, err := connect(config.Server, config.Port)
	if err != nil {
		log.Fatal(err)
	}

	// send messages: USER/NICK/JOIN
	conn.Write([]byte(msgUser(config.Nick)))
	conn.Write([]byte(msgNick(config.Nick)))
	conn.Write([]byte(msgJoin(config.Channel)))

	defer conn.Close()

	// define goroutines communicating via channel
	ch := make(chan string)
	go send(ch, conn)
	go receive(ch, conn)

	var input string
	fmt.Scanln(&input)
}
