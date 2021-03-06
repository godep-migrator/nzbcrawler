package ghost

import (
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

const (
	DAILY = "http://ghost-of-usenet.org/search.php?action=24h"
	LOGIN = "http://ghost-of-usenet.org/login.php"
)

func Redirect(req *http.Request, via []*http.Request) error {
	return errors.New("bla")
}

type GhostClient struct {
	User, Password string
	cookies        []*http.Cookie
	logged_in      bool
	dump           bool
}

func NewClient() (gc *GhostClient) {
	gc = &GhostClient{}
	gc.logged_in = false
	gc.dump = false
	return gc
}

func (g *GhostClient) SetAuth(user, password string) {
	g.User = user
	g.Password = password
}

func (g *GhostClient) SetDump(val bool) {
	g.dump = val
}

func (g GhostClient) IsLoggedIn() bool {
	return g.logged_in
}

func (g *GhostClient) getFirstTimeStuff() error {
	log.Infof("%s GET %v", TAG, LOGIN)
	log.Infof("%s getting cookies", TAG)
	client := &http.Client{}
	req, err := http.NewRequest("GET", LOGIN, nil)
	if err != nil {
		return err
	}

	req.Header.Add("Accept", "text/html, application/xhtml+xml, */*")
	req.Header.Add("Accept-Language", "de-DE")
	req.Header.Add("User-Agent", "Mozilla/5.0 (compatible; MSIE 10.0; Windows NT 6.2; WOW64; Trident/6.0)")
	req.Header.Add("Accept-Encoding", "gzip, deflate")
	req.Header.Add("Connection", "Keep-Alive")
	req.Header.Add("Host", "ghost-of-usenet.org")

	//connect to sUrl
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	g.cookies = resp.Cookies()

	return nil
}

// Logs into town.ag and returns the response cookies
func (g *GhostClient) Login() error {
	log.Infof("%s login process started", TAG)

	g.getFirstTimeStuff()

	param := url.Values{}
	param.Set("url", "index.php")
	param.Add("send", "send")
	param.Add("sid", "")
	param.Add("l_username", g.User)
	param.Add("l_password", g.Password)
	param.Add("submit", "Anmelden")

	client := &http.Client{}
	req, err := http.NewRequest("POST", LOGIN, strings.NewReader(param.Encode()))

	if err != nil {
		return err
	}

	log.Infof("%s POST %v", TAG, LOGIN)

	if g.cookies != nil {
		for _, cookie := range g.cookies {
			req.AddCookie(cookie)
		}
	}
	req.Header.Add("Accept", "text/html, application/xhtml+xml, */*")
	req.Header.Add("Referer", "http://ghost-of-usenet.org/index.php")
	req.Header.Add("Accept-Language", "de-DE")
	req.Header.Add("User-Agent", "Mozilla/5.0 (compatible; MSIE 10.0; Windows NT 6.2; WOW64; Trident/6.0)")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Accept-Encoding", "gzip, deflate")
	req.Header.Add("Host", "ghost-of-usenet.org")

	length := strconv.Itoa(len(param.Encode()))
	req.Header.Add("Content-Length", length)
	req.Header.Add("Connection", "Keep-Alive")
	req.Header.Add("Pragma", "no-cache")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	g.cookies = resp.Cookies()
	g.logged_in = true
	return nil
}

//http get to the given ressource
func (g *GhostClient) Get(sUrl string) (*http.Response, error) {
	log.Infof("%s GET %v", TAG, sUrl)

	client := &http.Client{}
	req, err := http.NewRequest("GET", sUrl, nil)
	if err != nil {
		log.Errorf("%s couldn't create Request to: %v", TAG, sUrl)
		return nil, err
	}

	req.Header.Add("Accept", "text/html, application/xhtml+xml, */*")
	req.Header.Add("Referer", "http://ghost-of-usenet.org/index.php")
	req.Header.Add("Accept-Language", "de-DE")
	req.Header.Add("User-Agent", "Mozilla/5.0 (compatible; MSIE 10.0; Windows NT 6.2; WOW64; Trident/6.0)")
	req.Header.Add("Accept-Encoding", "gzip, deflate")
	req.Header.Add("Host", "ghost-of-usenet.org")
	req.Header.Add("Connection", "Keep-Alive")

	if g.cookies != nil {
		for _, cookie := range g.cookies {
			req.AddCookie(cookie)
		}
	}

	//connect to sUrl
	resp, err := client.Do(req)
	if err != nil {
		log.Errorf("%s couldn't connect to: %v", TAG, sUrl)
		return nil, err
	}

	return resp, nil
}

//return the Daily url or "" if something went wrong
func (g *GhostClient) GetDailyUrl() (string, error) {
	client := &http.Client{
		CheckRedirect: Redirect,
	}
	log.Infof("%s GET url: %v", TAG, DAILY)
	req, err := http.NewRequest("GET", DAILY, nil)
	if err != nil {
		log.Errorf("%s %s", TAG, err.Error())
		return "", err
	}

	req.Header.Add("Accept", "text/html, application/xhtml+xml, */*")
	req.Header.Add("Referer", "http://ghost-of-usenet.org/index.php")
	req.Header.Add("Accept-Language", "de-DE")
	req.Header.Add("User-Agent", "Mozilla/5.0 (compatible; MSIE 10.0; Windows NT 6.2; WOW64; Trident/6.0)")
	req.Header.Add("Accept-Encoding", "gzip, deflate")
	req.Header.Add("Host", "ghost-of-usenet.org")
	req.Header.Add("Connection", "Keep-Alive")

	if g.cookies != nil {
		for _, cookie := range g.cookies {
			req.AddCookie(cookie)
		}
	}

	resp, err := client.Do(req)
	if resp == nil {
		return "", err
	}

	url, err := resp.Location()
	if err != nil {
		return "", err
	}
	return url.String(), nil

}
