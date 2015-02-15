package plugin

/*
usage: !ddg rehab in auckland
*/

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/dysfn/gerri/cmd"
	"github.com/dysfn/gerri/data"
	"net/url"
	"strings"
)

func escape(s string) string {
	return url.QueryEscape(strings.TrimSpace(s))
}

func ReplyDdg(pm data.Privmsg, config *data.Config) (string, error) {
	if len(pm.Message) >= 2 {
		term := escape(strings.Join(pm.Message[1:], " "))
		source := fmt.Sprintf("%s/html/?q=%s&kl=us-en&kp=-1", config.Ddg, term)

		doc, err := goquery.NewDocument(source)
		if err != nil {
			return "", nil
		}

		result := doc.Find("div.results_links").Not("div.web-result-sponsored") // excluded sponsored result
		title := strings.TrimSpace(result.Find("a.large").First().Text())
		url, _ := result.Find("a.large").Attr("href")
		url = strings.TrimSpace(url)

		if title != "" && url != "" {
			return cmd.Privmsg(pm.Target, fmt.Sprintf("%s (link: %s)", title, url)), nil
		} else {
			return cmd.PrivmsgAction(pm.Target, "zzzzz..."), nil
		}
	}
	return "", nil
}
