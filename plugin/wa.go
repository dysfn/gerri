package plugin

/*
usage: !wa Richard Stallman
*/

import (
	"encoding/xml"
	"fmt"
	"github.com/microamp/gerri/cmd"
	"github.com/microamp/gerri/data"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

func WAResult(config *data.Config, query string) (*data.WA, error) {
	var results *data.WA = &data.WA{}

	escaped_query := url.QueryEscape(strings.TrimSpace(query))
	resource := fmt.Sprintf("%s?input=%s&appid=%s", config.WaApi, escaped_query, config.WaKey)

	resp, err := http.Get(resource)
	if err != nil {
		return results, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return results, err
	}
	if err = xml.Unmarshal(body, results); err != nil {
		return results, err
	}
	return results, nil
}

func ReplyWA(pm data.Privmsg, config *data.Config) (string, error) {
	if len(pm.Message) >= 2 {
		query := strings.TrimSpace(strings.Join(pm.Message[1:], " "))
		results, err := WAResult(config, query)
		if err != nil {
			return "", err
		}

		var result_string string = ""
		for _, pod := range results.Pods {
			if pod.Title != "" && pod.Plaintext != "" {
				result_string += fmt.Sprintf("%s:{%s} ", pod.Title, pod.Plaintext)
			}
		}

		result_string = strings.Replace(string(result_string), "\n", " ", -1)
		words := strings.Split(result_string, " ")

		if len(words) > config.WaMaxWords {
			text := strings.Join(words[:config.WaMaxWords], " ")
			result_string = fmt.Sprintf("%s...", text)
		}

		return cmd.Privmsg(pm.Target, result_string), nil
	}
	return cmd.PrivmsgAction(pm.Target, "zzzzz..."), nil
}
