package main

import (
	"encoding/json"
	"errors"
	"flag"
	"html/template"
	"log"
	"net/http"

	"github.com/Xe/hideyhole-site/oauth2/discord"
	"github.com/facebookgo/flagconfig"
	"github.com/facebookgo/flagenv"
	"github.com/go-martini/martini"
	_ "github.com/lib/pq"
	"github.com/martini-contrib/csrf"
	moauth2 "github.com/martini-contrib/oauth2"
	"github.com/martini-contrib/sessions"
	"github.com/yosssi/ace"
	"github.com/yosssi/martini-acerender"
	"golang.org/x/oauth2"
	"gopkg.in/redis.v3"
)

func init() {
	moauth2.PathLogout = "/xXxXxXxXxlogout"
}

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
)

// DiscordUser is a user from discord/users/@me.
type DiscordUser struct {
	Username      string `json:"username"`
	Verified      bool   `json:"verified"`
	MfaEnabled    bool   `json:"mfa_enabled"`
	ID            string `json:"id" xorm:"pk"`
	Avatar        string `json:"avatar"`
	Discriminator string `json:"discriminator"`
	Email         string `json:"email"`
}

func (d *DiscordUser) Save(r *redis.Client) error {
	data, err := json.Marshal(d)
	if err != nil {
		return err
	}

	_, err = r.Set("discord:user:"+d.ID, string(data), 0).Result()
	if err != nil {
		return err
	}

	return nil
}

type Wrapper struct {
	Data    interface{}
	Session sessions.Session
}

type Site struct {
	redisClient *redis.Client
}

func (si *Site) getOwnDiscordUser(t moauth2.Tokens) (*DiscordUser, error) {
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
		return nil, errors.New("getOwnDiscordUser: " + resp.Status)
	}

	dUser := &DiscordUser{}

	err = json.NewDecoder(resp.Body).Decode(dUser)
	if err != nil {
		return nil, err
	}

	err = dUser.Save(si.redisClient)
	if err != nil {
		return nil, err
	}

	return dUser, nil
}

func (si *Site) populateInfo(s sessions.Session, t moauth2.Tokens) {
	otoken := s.Get("oauth2_token")
	if otoken == nil {
		return

	}

	uid := s.Get("uid")
	if uid == nil {
		dUser, err := si.getOwnDiscordUser(t)
		if err != nil {
			log.Printf("%v", err.Error())
			return
		}

		s.Set("uid", dUser.ID)
		s.Set("username", dUser.Username)
		s.Set("avatarhash", dUser.Avatar)
	}

	return
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

	si := &Site{
		redisClient: redis.NewClient(&redis.Options{
			Addr:     *redisHost,
			Password: *redisPassword,
			DB:       0,
		}),
	}

	_, err := si.redisClient.Ping().Result()
	if err != nil {
		log.Fatal(err)
	}

	m := martini.Classic()
	store := sessions.NewCookieStore([]byte(*cookieKey))

	m.Use(sessions.Sessions("cadeyforum", store))
	m.Use(acerender.Renderer(&acerender.Options{
		AceOptions: &ace.Options{
			BaseDir:       "views",
			DynamicReload: martini.Env == martini.Dev,
			FuncMap: template.FuncMap{
				"equals": func(a, b interface{}) bool {
					return a == b
				},
				"notequals": func(a, b interface{}) bool {
					return a != b
				},
			},
		},
	}))

	m.Use(moauth2.NewOAuth2Provider(discordOAuthClient))
	m.Use(csrf.Generate(&csrf.Options{
		Secret:     *cookieKey,
		SessionKey: *guildID,
		ErrorFunc: func(w http.ResponseWriter) {
			http.Error(w, "Bad request", http.StatusBadRequest)
		},
	}))
	m.Use(si.populateInfo)

	m.Get("/", si.getIndex)
	m.Get("/logout", si.logout)

	m.Get("/chat", moauth2.LoginRequired, si.getChat)
	m.Get("/profile/me", moauth2.LoginRequired, si.getMyProfile)

	m.RunOnAddr(":" + *port)
}
