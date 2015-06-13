package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	Web struct {
		Addr string `json:"addr"`
	} `json:"web"`

	Rpc struct {
		Addr string `json:"addr"`
		Crt  string `json:"crt"`
		Key  string `json:"key"`
	} `json:"rpc"`
}

func (c *Config) ReadFromFile(filename string) error {
	r, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer r.Close()

	return json.NewDecoder(r).Decode(c)
}
