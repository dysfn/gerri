package plugin

/*
usage: !beertime
*/

import (
	"fmt"
	"math"
	"strings"
	"time"
	"github.com/microamp/gerri/cmd"
	"github.com/microamp/gerri/data"
)

func timeDelta(weekday string, hour int, minute int) (string, error) {
	now := time.Now()
	wd := now.Weekday().String()
	if wd == weekday {
		y, m, d := now.Date()
		location := now.Location()

		beertime := time.Date(y, m, d, hour, minute, 0, 0, location)
		diff := beertime.Sub(now)

		if diff.Seconds() > 0 {
			return fmt.Sprintf("less than %d minute(s) to go...", int(math.Ceil(diff.Minutes()))), nil
		}
		return "it's beertime!", nil
	}
	return fmt.Sprintf("it's only %s...", strings.ToLower(wd)), nil
}

func ReplyBeertime(pm data.Privmsg, config *data.Config) (string, error) {
	td, err := timeDelta(config.Beertime.Day, config.Beertime.Hour, config.Beertime.Minute)
	if err != nil {
		return "", err
	}
	return cmd.Privmsg(pm.Target, td), nil
}
