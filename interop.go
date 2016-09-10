package main

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Xe/martini-oauth2"
)

// DiscordUser is a user from discord/users/@me.
type DiscordUser struct {
	Username string `json:"username"`
	ID       string `json:"id"`
	Avatar   string `json:"avatar" datastore:",noindex"`
	Email    string `json:"email"`
}

type DiscordUserGuild struct {
	Owner       bool   `json:"owner"`
	Permissions int    `json:"permissions"`
	Icon        string `json:"icon"`
	ID          string `json:"id"`
	Name        string `json:"name"`
}

func getOwnDiscordUser(t moauth2.Tokens) (*DiscordUser, error) {
	req, err := http.NewRequest("GET", "https://discordapp.com/api/users/@me", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+t.Access())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, errors.New("getOwnDiscordUser: " + resp.Status)
	}

	dUser := &DiscordUser{}

	err = json.NewDecoder(resp.Body).Decode(dUser)
	if err != nil {
		return nil, err
	}

	return dUser, nil
}

func getOwnDiscordGuilds(t moauth2.Tokens) ([]DiscordUserGuild, error) {
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

	guilds := []DiscordUserGuild{}

	err = json.NewDecoder(resp.Body).Decode(&guilds)
	if err != nil {
		return nil, err
	}

	return guilds, nil
}
