package plugin

/*
usage: !ping
*/

import (
	"github.com/dysfn/gerri/cmd"
	"github.com/dysfn/gerri/data"
)

func ReplyPing(pm data.Privmsg, config *data.Config) (string, error) {
	return cmd.PrivmsgAction(pm.Target, "meows"), nil
}
