package main

import (
	"github.com/go-ini/ini"
)

type Config struct{
	channelName string
	botUser string
	botToken string
}

func (c *Config) Load() error {
	cfg, err := ini.Load("config.ini")
	if err != nil {
		return err
	}

	c.channelName = cfg.Section("").Key("channelName").String()
	c.botUser = cfg.Section("").Key("botUser").String()
	c.botToken = cfg.Section("").Key("botToken").String()

	return nil
}