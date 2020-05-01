package main

import (
	"context"
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

func run() error {
	c, err := loadConfig("./config.yml")
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

	u, err := url.Parse(c.CallbackURL)
	if err != nil {
		return err
	}

	port, path := u.Port(), u.Path
	if port == "" {
		port = "80"
	}
	if path == "" {
		path = "/"
	}

	mux := http.NewServeMux()
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	ch := make(chan error, 1)

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
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
		close(ch)
	})

	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			ch <- err
		}
	}()
	if <-ch != nil {
		return err
	}

	if err = srv.Shutdown(context.TODO()); err != nil {
		return err
	}

	return nil
}

func main() {
	err := run()
	if err != nil {
		log.Fatal(err)
	}
}
