package plugin

/*
usage: !ud shane
*/

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/microamp/gerri/cmd"
	"github.com/microamp/gerri/data"
	"net/url"
	"strings"
)

func formatResult(result string, source string, limit int) string {
	first := strings.Split(strings.TrimSpace(result), "\n")[0]
	words := strings.Split(first, " ")

	var truncated string
	if len(words) > limit {
		truncated = fmt.Sprintf("%s...", strings.Join(words[:limit], " "))
	} else {
		truncated = first
	}

	if truncated != "" {
		return fmt.Sprintf("%s (source: %s)", truncated, source)
	}
	return ""
}

func ReplyUd(pm data.Privmsg, config *data.Config) (string, error) {
	if len(pm.Message) >= 2 {
		term := url.QueryEscape(strings.TrimSpace(strings.Join(pm.Message[1:], " ")))
		source := fmt.Sprintf("%s?term=%s", config.Ud, term)

		doc, err := goquery.NewDocument(source)
		if err != nil {
			return "", nil
		}

		result := formatResult(doc.Find("div.meaning").First().Text(), source, config.UdMaxWords)
		if result != "" {
			return cmd.Privmsg(pm.Target, result), nil
		} else {
			return cmd.PrivmsgAction(pm.Target, "zzzzz..."), nil
		}
	}
	return "", nil
}
