package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/dghubble/oauth1"
	"github.com/dghubble/oauth1/twitter"
	"gopkg.in/yaml.v2"
)

type config struct {
	ConsumerKey    string `yaml:"consumer_key"`
	ConsumerSecret string `yaml:"consumer_secret"`
	CallbackURL    string `yaml:"callback_url"`
}

func loadConfig(path string) (*config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	var c config
	if err = yaml.Unmarshal(b, &c); err != nil {
		return nil, err
	}

	return &c, nil
}

func getPort(u *url.URL) (string, error) {
	if u.Port() != "" {
		return u.Port(), nil
	}
	switch u.Scheme {
	case "http":
		return "80", nil
	case "https":
		return "443", nil
	default:
		return "", fmt.Errorf("invalid scheme %s", u.Scheme)
	}
}

func run() error {
	c, err := loadConfig("./config.yml")
	if err != nil {
		return err
	}

	u, err := url.Parse(c.CallbackURL)
	if err != nil {
		return err
	}

	o := oauth1.Config{
		ConsumerKey:    c.ConsumerKey,
		ConsumerSecret: c.ConsumerSecret,
		CallbackURL:    c.CallbackURL,
		Endpoint:       twitter.AuthorizeEndpoint,
	}

	rt, rs, err := o.RequestToken()
	if err != nil {
		return err
	}

	authorizationURL, err := o.AuthorizationURL(rt)
	fmt.Println(authorizationURL)

	callback := func(w http.ResponseWriter, r *http.Request) {
		_, verifier, err := oauth1.ParseAuthorizationCallback(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		at, as, err := o.AccessToken(rt, rs, verifier)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		io.WriteString(w, "ok\n")
		fmt.Println("Access Token: ", at)
		fmt.Println("Access Secret:", as)
	}

	port, err := getPort(u)
	if err != nil {
		return err
	}

	path := u.Path
	if path == "" {
		path = "/"
	}

	http.HandleFunc(path, callback)
	log.Fatal(http.ListenAndServe(":"+port, nil))

	return nil
}

func main() {
	err := run()
	if err != nil {
		log.Fatal(err)
	}
}
