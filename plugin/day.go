package plugin

/*
usage: !day
*/

import (
	"strings"
	"time"
	"github.com/microamp/gerri/cmd"
	"github.com/microamp/gerri/data"
)

func ReplyDay(pm data.Privmsg, config *data.Config) (string, error) {
	return cmd.Privmsg(pm.Target, strings.ToLower(time.Now().Weekday().String())), nil
}
