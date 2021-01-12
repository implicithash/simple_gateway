package config

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

// Config is a struct for json config
type Config struct {
	RequestPayload int `json:"request_payload"`
	RequestTimeout int `json:"request_timeout"`
}

var (
	// Cfg is a config proection
	Cfg = &Config{}
)

// Setup inits the app config
func Setup() error {
	data, err := ioutil.ReadFile("./src/config.json")
	if err != nil {
		log.Fatalln("cant read config file:", err)
		return err
	}

	err = json.Unmarshal(data, Cfg)
	if err != nil {
		log.Fatalln("cant parse config:", err)
		return err
	}
	return nil
}
