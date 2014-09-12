package plugin

/*
usage: !ping
*/

import (
	"github.com/microamp/gerri/cmd"
	"github.com/microamp/gerri/data"
)

func ReplyPing(pm data.Privmsg, config *data.Config) (string, error) {
	return cmd.PrivmsgAction(pm.Target, "meows"), nil
}
