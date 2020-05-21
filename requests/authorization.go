package requests

import (
	"encoding/json"
	"fmt"

	"github.com/andriiyaremenko/tinylog"
	"github.com/gocolly/colly"

	"net/http"
	"regexp"
)

var (
	srtRegex  = regexp.MustCompile(`\{"name":"srt","value":"\w+"\}`)
	rqidRegex = regexp.MustCompile(`\{"name":"rqid","value":"\w+"\}`)
	midRegex  = regexp.MustCompile(`"dfpmid":"(.+?)"`)
)

func NewAuthClient(loggerFactory tinylog.TinyLoggerFactory) Client {
	return &authClient{
		logger: loggerFactory.GetLogger("AuthClient"),
	}
}

type keyValue struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type authClient struct {
	logger tinylog.TinyLogger
}

// (ac *authClient) CallEndpoint(c *col.Collector, login, password string) *colly.Collector
func (ac *authClient) CallEndpoint(c *colly.Collector, args ...interface{}) *colly.Collector {
	if len(args) < 2 {
		ac.logger.Fatal("(ac *authClient) CallEndpoint(login, password string): to few arguments")
	}
	if len(args) > 2 {
		ac.logger.Fatal("(ac *authClient) CallEndpoint(login, password string): to much arguments")
	}
	login, password := args[0].(string), args[1].(string)

	c.OnResponse(func(r *colly.Response) {
		ac.logger.Debugf("%s: %s%s -> %d", r.Request.Method, r.Request.URL.Hostname(), r.Request.URL.RequestURI(), r.StatusCode)
	})

	c.OnError(func(r *colly.Response, err error) {
		if r.StatusCode == http.StatusMethodNotAllowed && r.Request.URL.Path == "/signin/s" {
			ac.logger.Errf("Status Code: %d; %v; Captcha check required", r.StatusCode, err)
			return
		}
		ac.logger.Fatalf("Status Code: %d; %v", r.StatusCode, err)
	})

	c.OnResponse(func(r *colly.Response) {
		switch r.Request.URL.Path {
		case "/":
			err := r.Request.Visit(`https://signin.ebay.com/ws/eBayISAPI.dll?SignIn&ru=https://www.ebay.com/`)
			if err != nil {
				ac.logger.Fatal(err)
			}
		case "/ws/eBayISAPI.dll":
			parseAndProceed(ac.logger, c, r, login, password)
		}
	})

	err := c.Visit("https://www.ebay.com/")
	if err != nil {
		ac.logger.Fatal(err)
	}
	return c.Clone()
}

func parseAndProceed(logger tinylog.TinyLogger, c *colly.Collector, r *colly.Response, login, password string) {
	body := r.Body
	srt := new(keyValue)
	err := json.Unmarshal(srtRegex.Find(body), srt)
	if err != nil {
		logger.Fatal(fmt.Sprintf("parseAndProceed: srt: %s", err))
	}
	rqid := new(keyValue)
	err = json.Unmarshal(rqidRegex.Find(body), rqid)
	if err != nil {
		logger.Fatal(fmt.Sprintf("parseAndProceed: rqid: %s", err))
	}
	midMatch := string(midRegex.Find(body))
	if midMatch == "" {
		logger.Fatal("parseAndProceed: dfpmid: failed to find dfpmid")
	}
	mid := midMatch[len(`"dfpmid":"`) : len(midMatch)-1]
	err = r.Request.Post("https://www.ebay.com/signin/s", map[string]string{
		"userid":                         login,
		"pass":                           password,
		"kmsi-unchecked":                 "1",
		"kmsi":                           "1",
		"pageType":                       "-1",
		"returnUrl":                      "https://www.ebay.com/",
		"srt":                            srt.Value,
		"rtmData":                        "PS=T.0",
		"rqid":                           rqid.Value,
		"lkdhjebhsjdhejdshdjchquwekguid": rqid.Value,
		"lastAttemptMethod":              "password",
		"showWebAuthnOptIn":              "1",
		"mid":                            mid,
		"isRecgUser":                     "false",
	})
	if err != nil {
		logger.Err(err)
	}
}
