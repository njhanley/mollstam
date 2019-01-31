package main

import (
	"flag"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"golang.org/x/sys/unix"
)

func updateChannel(dg *discordgo.Session, cfg *config, status, topic string) {
	if cfg.ChannelID == "" {
		return
	}
	channel := new(discordgo.ChannelEdit)
	if cfg.ChannelName != "" {
		channel.Name = cfg.ChannelName + "［" + status + "］"
	}
	if cfg.ChannelUpdateTopic {
		if topic == "" {
			channel.Topic = "\n" // workaround to clear channel topic
		} else {
			channel.Topic = topic
		}
	}
	if _, err := dg.ChannelEditComplex(cfg.ChannelID, channel); err != nil {
		log.Println("failed to update channel:", err)
	} else {
		log.Println("updated channel")
	}
}

func updateDiscord(dg *discordgo.Session, cfg *config, status *mcStatus) {
	if status == nil {
		updateChannel(dg, cfg, "offline", "")
	} else {
		players := make([]string, len(status.Players.Sample))
		for i, player := range status.Players.Sample {
			players[i] = player.Name
		}
		updateChannel(dg, cfg, strconv.Itoa(status.Players.Online), strings.Join(players, ", "))
	}
}

func notifyUser(dg *discordgo.Session, cfg *config) {
	if cfg.NotifyUserID == "" {
		return
	}
	if _, err := dg.ChannelMessageSend(cfg.ChannelID, "<@!"+cfg.NotifyUserID+"> "+cfg.NotifyMessage); err != nil {
		log.Println("failed to notify user:", err)
	} else {
		log.Println("notified user")
	}
}

func main() {
	configFilename := flag.String("c", "config.json", "config file location")
	flag.Parse()

	cfg, err := readConfig(*configFilename)
	if err != nil {
		fatal("failed to read config:", err)
	}

	pollingRate, err := time.ParseDuration(cfg.PollingRate)
	if err != nil {
		fatal("failed to parse PollingRate:", err)
	}

	timeout, err := time.ParseDuration(cfg.Timeout)
	if err != nil {
		fatal("failed to parse Timeout:", err)
	}

	dg, err := discordgo.New("Bot " + cfg.DiscordToken)
	if err != nil {
		fatal("failed to create session:", err)
	}

	err = dg.Open()
	if err != nil {
		fatal("failed to connect to Discord:", err)
	}
	defer dg.Close()

	go func() {
		var prevStatus *mcStatus
		var failedPings int
		for range time.Tick(pollingRate) {
			status, err := queryMinecraft(cfg.Address, timeout)
			if prevStatus != nil && err != nil {
				log.Println("failed to query server:", err)
			}
			if status == nil && prevStatus != nil || status != nil && (prevStatus == nil || status.Players.Online != prevStatus.Players.Online) {
				updateDiscord(dg, cfg, status)
			}
			if status == nil {
				failedPings++
				if failedPings == cfg.NotifyFailedPings {
					notifyUser(dg, cfg)
				}
			} else {
				failedPings = 0
			}
			prevStatus = status
		}
	}()

	await(unix.SIGINT, unix.SIGTERM)
}
