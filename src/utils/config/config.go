package config

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

type Config struct {
	RequestPayload int `json:"request_payload"`
	MaxRequestQty  int `json:"max_request_qty"`
	RequestTimeout int `json:"request_timeout"`
}

var (
	Cfg = &Config{}
)

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
