package plugin

/*
usage: !quote
usage: !quote 1
usage: !quote "i'm your father" - darth vader
*/

import (
	"log"
	"database/sql"
	"strings"
	"strconv"
	_ "github.com/mattn/go-sqlite3"
	"github.com/microamp/gerri/cmd"
	"github.com/microamp/gerri/data"
)

var QuoteDB *sql.DB = nil

func storeQuote(quoteStr string, sender string) (bool) {
	stmt, err := QuoteDB.Prepare(`
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
	row, err := QuoteDB.Query(`select text from quote order by random() limit 1;`)
	if err != nil || !row.Next() {
		return "", err
	}
	defer row.Close()

	var quoteStr string
	row.Scan(&quoteStr)
	return quoteStr, err
}

func getQuote(quoteID int64) (string, error) {
	stmt, err := QuoteDB.Prepare("select text from quote where id = (?);")
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

func ConnectQuoteDB(filename string) (*sql.DB) {
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

func ReplyQuote(pm data.Privmsg, config *data.Config) (string, error) {
	// Establish if we want to just return a single quote
	if len(pm.Message) == 2 {
		quoteID, err := strconv.ParseInt(pm.Message[1], 0, 0)
		if err == nil && quoteID > 0 {
			quoteStr, err := getQuote(quoteID)
			if err != nil {
				return "", err
			}
			return cmd.Privmsg(pm.Target, quoteStr), nil
		}
	}

	// Determine if this is a request for an existing quote, or attempting to
	// store a new one.
	msg := strings.Trim(strings.Join(pm.Message[1:], " "), " ")
	if strings.Count(msg, "\"") == 2 {
		if storeQuote(msg, pm.Source) {
			return cmd.Privmsg(pm.Target, "Got it!"), nil
		}
		return "", nil
	}

	// At this point the assumption is that we just want a random quote
	quoteStr, err := getRandomQuote()
	if err != nil {
		return "", err
	}
	return cmd.Privmsg(pm.Target, quoteStr), nil
}
