package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

func updateChannel(dg *discordgo.Session, cfg *config, status, topic string) {
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
		log.Printf("updated channel: %s players online\n", status)
	}
}

func notifyUser(dg *discordgo.Session, cfg *config) {
	if cfg.NotifyUserID == "" {
		return
	}
	if _, err := dg.ChannelMessageSend(cfg.ChannelID, "<@!"+cfg.NotifyUserID+"> "+cfg.NotifyMessage); err != nil {
		log.Println("failed to notify user:", err)
	} else {
		log.Printf("notified user: %s\n", cfg.NotifyMessage)
	}
}

func fatal(v ...interface{}) {
	fmt.Fprintln(os.Stderr, v...)
	os.Exit(1)
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
	if pollingRate < 5*time.Minute {
		log.Println("polling rate is less than 5 minutes; this may cause rate limit issues with Discord")
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
		var (
			failedPings int
			online      int = -1
			players     []string
		)
		for c := time.Tick(pollingRate); ; <-c {
			_online, _players, err := queryMinecraft(cfg.Address, timeout)
			if err != nil {
				log.Println("failed to query server:", err)
				failedPings++
				if failedPings == cfg.NotifyFailedPings {
					updateChannel(dg, cfg, "offline", "")
					notifyUser(dg, cfg)
				}
				continue
			} else {
				failedPings = 0
			}
			if _online != online || reflect.DeepEqual(_players, players) {
				online, players = _online, _players
				topic := strings.Join(players, ", ")
				if 0 < len(players) && len(players) < online {
					topic += ", ..."
				}
				updateChannel(dg, cfg, strconv.Itoa(online), topic)
			}
		}
	}()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, os.Interrupt)
	<-sc
}
