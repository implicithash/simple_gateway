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
	MaxQueueSize   int `json:"max_queue_size"`
	IncomingReqQty int `json:"incoming_req_qty"`
	OutgoingReqQty int `json:"outgoing_req_qty"`
}

var (
	// Cfg is a config projection
	Cfg = &Config{}
)

// Setup inits the app config
func Setup() error {
	data, err := ioutil.ReadFile("./config.json")
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
