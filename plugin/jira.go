package plugin

/*
usage: !jira
usage: !jira taskname
*/

import (
	"github.com/dysfn/gerri/cmd"
	"github.com/dysfn/gerri/data"
	"strings"
)

func ReplyJira(pm data.Privmsg, config *data.Config) (string, error) {
	msg := strings.Join(pm.Message[1:], " ")
	if strings.TrimSpace(msg) != "" {
		return cmd.Privmsg(pm.Target, config.Jira+"/browse/"+strings.ToUpper(msg)), nil
	}
	return cmd.Privmsg(pm.Target, config.Jira), nil
}
