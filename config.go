package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
)

type config struct {
	Address     string `json:"address"`      // server address and port (default: "127.0.0.1:25565")
	PollingRate string `json:"polling_rate"` // duration between pings (default: "5m")
	Timeout     string `json:"timeout"`      // duration to wait when connecting to the server (default: "5s")

	DiscordToken       string `json:"discord_token"`        // bot token (required)
	ChannelID          string `json:"channel_id"`           // channel used for player count, list, and notifications (required)
	ChannelName        string `json:"channel_name"`         // name used when updating channel (omit to disable updating channel name)
	ChannelUpdateTopic bool   `json:"channel_update_topic"` // update channel topic with player list (default: false)
	NotifyUserID       string `json:"notify_user_id"`       // ID of user to notify when server is unreachable (omit to disable notifying user)
	NotifyFailedPings  int    `json:"notify_failed_pings"`  // failed pings before user is notified (default: 5)
	NotifyMessage      string `json:"notify_message"`       // message sent to user when server is unreachable (default: "The server appears to be offline.")
}

func readConfig(filename string) (*config, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	cfg := &config{
		Address:           "127.0.0.1:25565",
		PollingRate:       "5m",
		Timeout:           "5s",
		NotifyFailedPings: 5,
		NotifyMessage:     "The server appears to be offline.",
	}
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	if cfg.DiscordToken == "" {
		return nil, errors.New(`"discord_token" is required`)
	}
	if cfg.ChannelID == "" {
		return nil, errors.New(`"channel_id" is required`)
	}
	return cfg, nil
}
