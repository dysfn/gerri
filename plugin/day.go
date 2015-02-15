package plugin

/*
usage: !day
*/

import (
	"github.com/dysfn/gerri/cmd"
	"github.com/dysfn/gerri/data"
	"strings"
	"time"
)

func ReplyDay(pm data.Privmsg, config *data.Config) (string, error) {
	return cmd.Privmsg(pm.Target, strings.ToLower(time.Now().Weekday().String())), nil
}
