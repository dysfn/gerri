package plugin

/*
usage: !slap someone
*/

import (
	"fmt"
	"math/rand"
	"strings"
	"github.com/microamp/gerri/cmd"
	"github.com/microamp/gerri/data"
)

func slapAction(target string) (string, error) {
	actions := []string {
		"slaps", "kicks", "destroys", "annihilates", "punches",
		"roundhouse kicks", "rusty hooks", "pwns", "owns"}
	if strings.TrimSpace(target) != "" {
		selected_action := actions[rand.Intn(len(actions))]
		return fmt.Sprintf("%s %s", selected_action, target), nil
	}
	return "zzzzz...", nil
}

func ReplySlap(pm data.Privmsg, config *data.Config) (string, error) {
	slap, err := slapAction(strings.Join(pm.Message[1:], " "))
	if err != nil {
		return "", err
	}
	return cmd.PrivmsgAction(pm.Target, slap), nil
}
