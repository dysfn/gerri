package plugin

/*
usage: !ver
usage: !version
*/

import (
	"fmt"
	"github.com/dysfn/gerri/cmd"
	"github.com/dysfn/gerri/data"
)

const (
	VERSION = "0.3.4"
)

func ReplyVer(pm data.Privmsg, config *data.Config) (string, error) {
	return cmd.Privmsg(pm.Target, fmt.Sprintf("gerri version: %s", VERSION)), nil
}
