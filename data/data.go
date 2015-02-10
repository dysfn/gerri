package data

/*
(data) struct(ures)
*/

type Config struct {
	Server      string
	Port        string
	Nick        string
	Channel     string
	WikMaxWords int
	Ud          string
	UdMaxWords  int
	Giphy       string
	GiphyApi    string
	Ddg         string
	DdgApi      string
	GiphyKey    string
	Jira        string
	Beertime    Beertime
	AdviceApi   string
	QuoteDB     string
	SlapActions []string
	PyApi       string
	WaApi       string
	WaKey       string
	WaMaxWords  int
}

type Beertime struct {
	Day    string
	Hour   int
	Minute int
}

type Privmsg struct {
	Source  string
	Target  string
	Message []string
}

type DuckDuckGo struct {
	AbstractText string
	AbstractURL  string
}

type GIF struct {
	ID string
}

type Giphy struct {
	Data []GIF
}

type Advice struct {
	Quote string `xml:"quote"`
}

type Pod struct {
	Title     string `xml:"title,attr"`
	Plaintext string `xml:"subpod>plaintext"`
}

type WA struct {
	Pods []Pod `xml:"pod"`
}
