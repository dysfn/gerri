package cmd

/*
IRC commands
*/

import "fmt"

const (
	USER    = "USER"
	NICK    = "NICK"
	JOIN    = "JOIN"
	PING    = "PING"
	PONG    = "PONG"
	PRIVMSG = "PRIVMSG"
	ACTION  = "ACTION"
	SUFFIX  = "\r\n"
)

func User(nick string) string {
	return USER + " " + nick + " 8 * :" + nick + SUFFIX
}

func Nick(nick string) string {
	return NICK + " " + nick + SUFFIX
}

func Join(channel string) string {
	return JOIN + " " + channel + SUFFIX
}

func Pong(host string) string {
	return PONG + " :" + host + SUFFIX
}

func Privmsg(receiver string, msg string) string {
	return PRIVMSG + " " + receiver + " :" + msg + SUFFIX
}

func PrivmsgAction(receiver string, msg string) string {
	return fmt.Sprintf("%s %s :\001%s %s\001%s", PRIVMSG, receiver, ACTION, msg, SUFFIX)
}
