// Read and setting application config
package config

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"chromedp"
	"time"
)

const (
	ProductEnvironment     = "prod"
	DevelopmentEnvironment = "dev"
)

type Api struct {
	Prod string
	Dev  string
}

type Config struct {
	Debug           bool
	Env             string
	ApiEndpoint     string
	ApiConfig       Api
	DataSource      string
	IntervalSeconds time.Duration
	Chromedp        chromedp.ChromeDp
}

func NewConfig() *Config {
	cfg := &Config{}
	str, err := ioutil.ReadFile("config/config.json")
	if err != nil {
		log.Fatalln("Read config file error:", err)
	}

	json.Unmarshal(str, cfg)
	if cfg.Env == ProductEnvironment {
		cfg.ApiEndpoint = cfg.ApiConfig.Prod
	} else {
		cfg.ApiEndpoint = cfg.ApiConfig.Dev
	}

	return cfg
}
