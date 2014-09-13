package plugin

/*
usage: !ver
usage: !version
*/

import (
	"fmt"
	"github.com/microamp/gerri/cmd"
	"github.com/microamp/gerri/data"
)

const (
	VERSION = "0.3.2"
)

func ReplyVer(pm data.Privmsg, config *data.Config) (string, error) {
	return cmd.Privmsg(pm.Target, fmt.Sprintf("gerri version: %s", VERSION)), nil
}
