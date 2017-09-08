package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/BurntSushi/toml"
)

func main() {
	config, err := ParseConfigFile("config.toml")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println(config.Host)
	fmt.Println(config.Port)
}

type Config struct {
	Host         string `toml:"mysql_host"`
	Port         int    `toml:"mysql_port"`
	User         string `toml:"user"`
	Password     string `toml:"password"`
	DBName       string `toml:"db_name"`
	MaxIdleConns int    `toml:"max_idle_conns"`
}

func ParseConfigFile(filename string) (*Config, error) {
	var cfg Config
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	_, err = toml.Decode(string(data), &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}
