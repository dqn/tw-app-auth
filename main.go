package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
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

	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET(u.Path, hello)

	port, err := getPort(u)
	if err != nil {
		return err
	}

	e.Logger.Fatal(e.Start(":" + port))

	return nil
}

func main() {
	err := run()
	if err != nil {
		log.Fatal(err)
	}
}

func hello(c echo.Context) error {
	return c.String(http.StatusOK, "Hello, World!")
}
