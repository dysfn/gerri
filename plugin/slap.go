package plugin

/*
usage: !slap someone
*/

import (
	"fmt"
	"github.com/dysfn/gerri/cmd"
	"github.com/dysfn/gerri/data"
	"math/rand"
	"regexp"
	"strings"
)

func slapAction(target string, source_nick string, config *data.Config) (string, error) {
	selected_action := config.SlapActions[rand.Intn(len(config.SlapActions))]
	if strings.TrimSpace(target) != "" && strings.TrimSpace(target) != config.Nick {
		return fmt.Sprintf("%s %s", selected_action, target), nil
	} else {
		return fmt.Sprintf("%s %s", selected_action, source_nick), nil
	}
	return "zzzzz...", nil
}

func Nick(source string) (string, error) {
	/*
	   Extracts nick from the IRC source data.
	*/
	r, err := regexp.Compile(".(?P<nick>.+)!.*")
	if err != nil {
		return "", err
	}
	return r.FindAllStringSubmatch(source, 1)[0][1], nil
}

func ReplySlap(pm data.Privmsg, config *data.Config) (string, error) {
	source_nick, err := Nick(pm.Source)
	slap, err := slapAction(
		strings.Join(pm.Message[1:], " "), source_nick, config)
	if err != nil {
		return "", err
	}
	return cmd.PrivmsgAction(pm.Target, slap), nil
}
