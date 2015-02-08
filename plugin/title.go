package plugin

/*
usage: !title http://example.com
*/

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/microamp/gerri/cmd"
	"github.com/microamp/gerri/data"
	"strings"
)

func ReplyTitle(pm data.Privmsg, config *data.Config) (string, error) {
	if len(pm.Message) == 2 {
		url := strings.TrimSpace(pm.Message[1])
		if url != "" {
			doc, err := goquery.NewDocument(url)
			if err != nil {
				return "", err
			}

			return cmd.Privmsg(pm.Target, doc.Find("title").Text()), nil
		}
	}
	return "", nil
}
