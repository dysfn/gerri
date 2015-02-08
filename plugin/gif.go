package plugin

/*
usage: !gif
usage: !gif kittens
*/

import (
	"encoding/json"
	"fmt"
	"github.com/microamp/gerri/cmd"
	"github.com/microamp/gerri/data"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
)

func searchGiphy(term string, config *data.Config) (*data.Giphy, error) {
	var giphy *data.Giphy = &data.Giphy{}

	if term == "" {
		term = "cat"
	}
	encoded := url.QueryEscape(term)
	resource := fmt.Sprintf("%s/v1/gifs/search?api_key=%s&q=%s", config.GiphyApi, config.GiphyKey, encoded)

	resp, err := http.Get(resource)
	if err != nil {
		return giphy, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return giphy, err
	}
	if err = json.Unmarshal(body, giphy); err != nil {
		return giphy, err
	}
	return giphy, nil
}

func ReplyGIF(pm data.Privmsg, config *data.Config) (string, error) {
	msg := strings.Join(pm.Message[1:], " ")
	giphy, err := searchGiphy(msg, config)
	if err != nil {
		return "", err
	}
	if len(giphy.Data) > 0 {
		m := fmt.Sprintf("%s/media/%s/giphy.gif", config.Giphy, giphy.Data[rand.Intn(len(giphy.Data))].ID)
		return cmd.Privmsg(pm.Target, m), nil
	}
	return cmd.PrivmsgAction(pm.Target, "zzzzz..."), nil
}
