package plugin

/*
usage:
    !time or !time local
    !time Colombo
    !time NZDT
*/

import (
	"fmt"
	"github.com/dysfn/gerri/cmd"
	"github.com/dysfn/gerri/data"
	"strings"
)

const (
	TIME       = "time"
	LOCAL_TIME = "local time"
	RESULT     = "Result"
)

func ReplyTime(pm data.Privmsg, config *data.Config) (string, error) {
	query := ""
	if len(pm.Message) >= 2 {
		query = strings.TrimSpace(strings.Join(pm.Message[1:], " "))
	}

	if query == "" {
		query = LOCAL_TIME
	}

	if !strings.Contains(query, TIME) {
		query = fmt.Sprintf("%s %s", query, TIME)
	}

	results, err := WAResult(config, query)
	if err != nil {
		return "", err
	}

	for _, pod := range results.Pods {
		if pod.Title == RESULT && pod.Plaintext != "" {
			return cmd.Privmsg(pm.Target, pod.Plaintext), nil
		}
	}
	return cmd.PrivmsgAction(pm.Target, "zzzzz..."), nil
}
