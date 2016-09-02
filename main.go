package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/Xe/hideyhole-site/oauth2/discord"
	"github.com/facebookgo/flagconfig"
	"github.com/facebookgo/flagenv"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/csrf"
	moauth2 "github.com/martini-contrib/oauth2"
	"github.com/martini-contrib/sessions"
	"golang.org/x/oauth2"
	"gopkg.in/redis.v3"
)

var (
	clientID      = flag.String("discord-client-id", "", "discord oauth client id")
	clientSecret  = flag.String("discord-client-secret", "", "discord oauth client secret")
	redisHost     = flag.String("redis-host", "127.0.0.1:6379", "redis server to use for memory")
	redisPassword = flag.String("redis-password", "", "redis serer password")
	pqURL         = flag.String("database-url", "", "database URL")
	port          = flag.String("port", "3093", "TCP port to listen on for HTTP requests")
	guildID       = flag.String("guild-id", "", "guild ID for allowing membership")
	cookieKey     = flag.String("cookie-key", "", "random cookie key")
	salt          = flag.String("salt", "", "salt for any passwords or crypto stuff")

	discordOAuthClient *oauth2.Config
	redisClient        *redis.Client
)

type DiscordUser struct {
	Username      string `json:"username"`
	Verified      bool   `json:"verified"`
	MfaEnabled    bool   `json:"mfa_enabled"`
	ID            string `json:"id"`
	Avatar        string `json:"avatar"`
	Discriminator string `json:"discriminator"`
	Email         string `json:"email"`
}

func getDiscordUser(t moauth2.Tokens) (*DiscordUser, error) {
	req, err := http.NewRequest("GET", "https://discordapp.com/api/users/@me", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+t.Access())

	c := &http.Client{}
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, errors.New("getDiscordUser: " + resp.Status)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	dUser := &DiscordUser{}

	err = json.NewDecoder(bytes.NewBuffer(data)).Decode(dUser)
	if err != nil {
		return nil, err
	}

	log.Printf("%#v", dUser)

	recorded, err := redisClient.Exists("discord:user:" + dUser.ID).Result()
	if err != nil {
		return nil, err
	}

	if !recorded {
		_, err := redisClient.Set("discord:user:"+dUser.ID, string(data), 0).Result()
		if err != nil {
			return nil, err
		}
	}

	return dUser, nil
}

func main() {
	flagenv.Parse()
	flag.Parse()
	flagconfig.Parse()

	discordOAuthClient = &oauth2.Config{
		ClientID:     *clientID,
		ClientSecret: *clientSecret,
		Endpoint:     discord.Endpoint,
		Scopes:       []string{"identify", "email", "guilds"},
		RedirectURL:  "http://greedo.xeserv.us:3093" + moauth2.PathCallback,
	}

	redisClient = redis.NewClient(&redis.Options{
		Addr:     *redisHost,
		Password: *redisPassword,
		DB:       0,
	})

	_, err := redisClient.Ping().Result()
	if err != nil {
		log.Fatal(err)
	}

	m := martini.Classic()
	store := sessions.NewCookieStore([]byte(*cookieKey))
	m.Use(sessions.Sessions("cadeyforum", store))

	m.Use(moauth2.NewOAuth2Provider(discordOAuthClient))
	// m.Use(moauth2.LoginRequired)

	m.Use(csrf.Generate(&csrf.Options{
		Secret:     *cookieKey,
		SessionKey: *guildID,
		ErrorFunc: func(w http.ResponseWriter) {
			http.Error(w, "Bad request", http.StatusBadRequest)
		},
	}))

	m.Use(
		func(s sessions.Session, t moauth2.Tokens, w http.ResponseWriter, r *http.Request) {
			otoken := s.Get("oauth2_token")
			if otoken == nil {
				http.Redirect(w, r, "/login", 302)
				return
			}

			uid := s.Get("uid")
			if uid == nil {
				dUser, err := getDiscordUser(t)
				if err != nil {
					http.Error(w, err.Error(), 500)
					return
				}

				s.Set("uid", dUser.ID)
			}
		},
	)

	m.RunOnAddr(":" + *port)
}
