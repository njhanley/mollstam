package main

import (
	"encoding/json"

	"github.com/bwmarrin/discordgo"
)

// copied from github.com/bwmarrin/discordgo/restapi.go
func unmarshal(data []byte, v interface{}) error {
	err := json.Unmarshal(data, v)
	if err != nil {
		return discordgo.ErrJSONUnmarshal
	}

	return nil
}

type channelEdit struct {
	Name                 string                           `json:"name,omitempty"`
	Topic                string                           `json:"topic"`
	NSFW                 bool                             `json:"nsfw,omitempty"`
	Position             int                              `json:"position"`
	Bitrate              int                              `json:"bitrate,omitempty"`
	UserLimit            int                              `json:"user_limit,omitempty"`
	PermissionOverwrites []*discordgo.PermissionOverwrite `json:"permission_overwrites,omitempty"`
	ParentID             string                           `json:"parent_id,omitempty"`
}

// same as discordgo.Session.ChannelEditComplex
// except the channel topic can be empty
func channelEditComplex(s *discordgo.Session, channelID string, data *channelEdit) (st *discordgo.Channel, err error) {
	body, err := s.RequestWithBucketID("PATCH", discordgo.EndpointChannel(channelID), data, discordgo.EndpointChannel(channelID))
	if err != nil {
		return
	}

	err = unmarshal(body, &st)
	return
}
