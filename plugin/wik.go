package plugin

/*
usage: !wik whatever
*/

import (
	"encoding/json"
	"fmt"
	"github.com/dysfn/gerri/cmd"
	"github.com/dysfn/gerri/data"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

func queryDuckDuckGo(term string, config *data.Config) (*data.DuckDuckGo, error) {
	var ddg *data.DuckDuckGo = &data.DuckDuckGo{}

	encoded := url.QueryEscape(term)
	resource := fmt.Sprintf("%s?format=json&q=%s", config.DdgApi, encoded)

	resp, err := http.Get(resource)
	if err != nil {
		return ddg, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ddg, err
	}
	if err = json.Unmarshal(body, ddg); err != nil {
		return ddg, err
	}
	return ddg, nil
}

func ReplyWik(pm data.Privmsg, config *data.Config) (string, error) {
	msg := strings.Join(pm.Message[1:], " ")
	if strings.TrimSpace(msg) != "" {
		ddg, err := queryDuckDuckGo(msg, config)
		if err != nil {
			return "", err
		}
		if ddg.AbstractText != "" && ddg.AbstractURL != "" {
			words := strings.Split(ddg.AbstractText, " ")
			var m string
			if len(words) > config.WikMaxWords {
				text := strings.Join(words[:config.WikMaxWords], " ")
				m = fmt.Sprintf("%s... (source: %s)", text, ddg.AbstractURL)
			} else {
				m = fmt.Sprintf("%s (source: %s)", ddg.AbstractText, ddg.AbstractURL)
			}
			return cmd.Privmsg(pm.Target, m), nil
		}
		return cmd.PrivmsgAction(pm.Target, "zzzzz..."), nil
	}
	return "", nil
}
