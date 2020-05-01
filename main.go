package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/dghubble/oauth1"
	"github.com/dghubble/oauth1/twitter"
)

func run() error {
	if len(os.Args) != 4 {
		fmt.Println("Usage: go run main.go <consumer key> <consumer secret> <callback url>")
		os.Exit(2)
	}

	ck, cs, callbackURL := os.Args[1], os.Args[2], os.Args[3]

	o := oauth1.Config{
		ConsumerKey:    ck,
		ConsumerSecret: cs,
		CallbackURL:    callbackURL,
		Endpoint:       twitter.AuthorizeEndpoint,
	}

	rt, rs, err := o.RequestToken()
	if err != nil {
		return err
	}

	authorizationURL, err := o.AuthorizationURL(rt)
	fmt.Println(authorizationURL)

	u, err := url.Parse(callbackURL)
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

		fmt.Println("access token: ", at)
		fmt.Println("access secret:", as)

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
