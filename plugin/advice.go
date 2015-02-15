package plugin

/*
usage: !advice
*/

import (
	"encoding/xml"
	"fmt"
	"github.com/dysfn/gerri/cmd"
	"github.com/dysfn/gerri/data"
	"io/ioutil"
	"net/http"
	"strings"
)

func getAdvice(config *data.Config) (*data.Advice, error) {
	var advice *data.Advice = &data.Advice{}

	resource := fmt.Sprintf(config.AdviceApi)

	resp, err := http.Get(resource)
	if err != nil {
		return advice, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return advice, err
	}
	if err = xml.Unmarshal(body, advice); err != nil {
		return advice, err
	}
	return advice, nil
}

func ReplyAdvice(pm data.Privmsg, config *data.Config) (string, error) {
	advice, err := getAdvice(config)
	if err != nil {
		return "", err
	}
	if advice.Quote != "" {
		return cmd.Privmsg(pm.Target, strings.ToLower(advice.Quote)), nil
	}
	return cmd.PrivmsgAction(pm.Target, "zzzzz..."), nil
}
