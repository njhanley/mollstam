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
	channel := new(channelEdit)
	if cfg.ChannelName != "" {
		channel.Name = cfg.ChannelName + "［" + status + "］"
	}
	channel.Topic = topic
	if _, err := channelEditComplex(dg, cfg.ChannelID, channel); err != nil {
		log.Println("failed to update channel:", err)
	} else {
		log.Println("updated channel")
	}
}

func updateDiscord(dg *discordgo.Session, cfg *config, status *mcStatus) {
	if status == nil {
		updateChannel(dg, cfg, "offline", "")
		if cfg.NotifyUserID == "" {
			return
		}
		if _, err := dg.ChannelMessageSend(cfg.ChannelID, "<@!"+cfg.NotifyUserID+"> "+cfg.NotifyMessage); err != nil {
			log.Println("failed to notify user:", err)
		} else {
			log.Println("notified user")
		}
	} else {
		players := make([]string, len(status.Players.Sample))
		for i, player := range status.Players.Sample {
			players[i] = player.Name
		}
		updateChannel(dg, cfg, strconv.Itoa(status.Players.Online), strings.Join(players, ", "))
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
		fatal(err)
	}

	timeout, err := time.ParseDuration(cfg.PollingRate)
	if err != nil {
		fatal(err)
	}

	dg, err := discordgo.New("Bot " + cfg.DiscordToken)
	if err != nil {
		fatal(err)
	}

	err = dg.Open()
	if err != nil {
		fatal(err)
	}
	defer dg.Close()

	go func() {
		var prevStatus *mcStatus
		for range time.Tick(pollingRate) {
			status, err := queryMinecraft(cfg.Address, timeout)
			if prevStatus != nil && err != nil {
				log.Println("failed to query server:", err)
			}
			if status == nil && prevStatus != nil || status != nil && (prevStatus == nil || status.Players.Online != prevStatus.Players.Online) {
				updateDiscord(dg, cfg, status)
			}
			prevStatus = status
		}
	}()

	await(unix.SIGINT, unix.SIGTERM)
}
