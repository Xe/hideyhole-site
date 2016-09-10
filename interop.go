package main

import (
	"encoding/json"
	"net/http"

	"github.com/Xe/martini-oauth2"
)

type DiscordUserGuild struct {
	Owner       bool   `json:"owner"`
	Permissions int    `json:"permissions"`
	Icon        string `json:"icon"`
	ID          string `json:"id"`
	Name        string `json:"name"`
}

func getOwnDiscordGuilds(t moauth2.Tokens) ([]*DiscordUserGuild, error) {
	req, err := http.NewRequest("GET", "https://discordapp.com/api/users/@me/guilds", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+t.Access())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	guilds := []*DiscordUserGuild{}

	err = json.NewDecoder(resp.Body).Decode(guilds)
	if err != nil {
		return nil, err
	}

	return guilds, nil
}
