package discordwidget

import (
	"encoding/json"
	"net/http"
)

type Guild struct {
	Channels  []Channel `json:"channels"`
	InviteURL string    `json:"instant_invite"`
	ID        string    `json:"id"`
	Members   []Member  `json:"members"`
	Name      string    `json:"name"`
}

type Channel struct {
	Position int    `json:"position"`
	ID       string `json:"id"`
	Name     string `json:"name"`
}

type Game struct {
	Name string `json:"name"`
}

type Member struct {
	Username      string `json:"username"`
	Status        string `json:"status"`
	AvatarURL     string `json:"avatar_url"`
	Avatar        string `json:"avatar"`
	Discriminator string `json:"discriminator"`
	ID            string `json:"id"`
	Game          *Game  `json:"game,omitempty"`
	Mute          bool   `json:"mute,omitempty"`
	Suppress      bool   `json:"suppress,omitempty"`
	Deaf          bool   `json:"deaf,omitempty"`
	ChannelID     string `json:"channel_id,omitempty"`
	SelfDeaf      bool   `json:"self_deaf,omitempty"`
	SelfMute      bool   `json:"self_mute,omitempty"`
	Nick          string `json:"nick,omitempty"`
}

func (m Member) GetGame() string {
	if m.Game == nil {
		return ""
	}

	return m.Game.Name
}

func GetGuild(guildID string) (*Guild, error) {
	result := &Guild{}

	resp, err := http.Get("https://discordapp.com/api/servers/" + guildID + "/widget.json")
	if err != nil {
		return nil, err
	}

	err = json.NewDecoder(resp.Body).Decode(result)
	if err != nil {
		return nil, err
	}

	return result, nil
}
