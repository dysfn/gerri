package plugin

/*
usage: !py "hello%20world".title()
*/

import (
	"fmt"
	"github.com/dysfn/gerri/cmd"
	"github.com/dysfn/gerri/data"
	"io/ioutil"
	"net/http"
	"strings"
)

func PyResult(config *data.Config, statement string) (string, error) {
	var result string

	resource := fmt.Sprintf("%s/%s", config.PyApi, statement)

	resp, err := http.Get(resource)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	result = strings.Replace(string(body), "\n", " ", -1)
	result = strings.TrimSpace(result)

	if err != nil {
		return result, err
	}
	return result, nil
}

func ReplyPy(pm data.Privmsg, config *data.Config) (string, error) {
	if len(pm.Message) >= 2 {
		statement := strings.TrimSpace(strings.Join(pm.Message[1:], " "))
		result, err := PyResult(config, statement)
		if err != nil {
			return "", err
		}
		if result != "" {
			return cmd.Privmsg(pm.Target, result), nil
		}
		return cmd.PrivmsgAction(pm.Target, "zzzzz..."), nil
	}
	return "zzzzz...", nil
}
