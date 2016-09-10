package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"html/template"
	"log"
	"net/http"
	"net/http/pprof"

	"github.com/Xe/hideyhole-site/oauth2/discord"
	"github.com/Xe/martini-oauth2"
	"github.com/facebookgo/flagconfig"
	"github.com/facebookgo/flagenv"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/csrf"
	"github.com/martini-contrib/sessions"
	"github.com/yosssi/ace"
	"github.com/yosssi/martini-acerender"
	"golang.org/x/oauth2"
)

var (
	clientID        = flag.String("discord-client-id", "", "discord oauth client id")
	clientSecret    = flag.String("discord-client-secret", "", "discord oauth client secret")
	googleProjectID = flag.String("google-project-id", "", "google project ID")
	port            = flag.String("port", "3093", "TCP port to listen on for HTTP requests")
	guildID         = flag.String("guild-id", "", "guild ID for allowing membership")
	cookieKey       = flag.String("cookie-key", "", "random cookie key")
	salt            = flag.String("salt", "", "salt for any passwords or crypto stuff")
	debug           = flag.Bool("debug", false, "add /debug routes? pprof, etc.")

	discordOAuthClient *oauth2.Config
)

// DiscordUser is a user from discord/users/@me.
type DiscordUser struct {
	Username string `json:"username"`
	ID       string `json:"id"`
	Avatar   string `json:"avatar" datastore:",noindex"`
	Email    string `json:"email"`
}

type Wrapper struct {
	Data    interface{}
	Session sessions.Session
}

type Site struct {
	db *Database
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

	return dUser, nil
}

func (si *Site) populateInfo(s sessions.Session, t moauth2.Tokens) error {
	otoken := s.Get("oauth2_token")
	if otoken == nil {
		return nil
	}

	uid := s.Get("uid")
	if uid == nil {
		dUser, err := si.getOwnDiscordUser(t)
		if err != nil {
			return err
		}

		s.Set("uid", dUser.ID)
		s.Set("username", dUser.Username)
		s.Set("avatarhash", dUser.Avatar)

		err = si.db.PutUser(context.Background(), dUser)
		if err != nil {
			return err
		}
	}

	return nil
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

	si := &Site{}

	db, err := initDB()
	if err != nil {
		log.Fatal(err)
	}

	si.db = db

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

	m.Use(moauth2.NewOAuth2Provider(discordOAuthClient, si.populateInfo))
	m.Use(csrf.Generate(&csrf.Options{
		Secret:     *cookieKey,
		SessionKey: *guildID,
		ErrorFunc: func(w http.ResponseWriter) {
			http.Error(w, "Bad request", http.StatusBadRequest)
		},
	}))
	m.Use(si.populateInfo)

	m.Get("/", si.getIndex)
	m.Get("/chat", si.getChat)
	m.Get("/health", si.getHealth)

	m.Get("/profile/me", moauth2.LoginRequired, si.getMyProfile)
	m.Get("/profile/:id", moauth2.LoginRequired, si.getUserByID)

	if *debug {
		log.Printf("Adding /debug routes")
		if martini.Env == martini.Prod {
			log.Printf("The pprof routes are enabled in production!!! Please act with care.")
		}

		m.Get("/debug/pprof", pprof.Index)
		m.Get("/debug/pprof/cmdline", pprof.Cmdline)
		m.Get("/debug/pprof/profile", pprof.Profile)
		m.Get("/debug/pprof/symbol", pprof.Symbol)
		m.Post("/debug/pprof/symbol", pprof.Symbol)
		m.Get("/debug/pprof/block", pprof.Handler("block").ServeHTTP)
		m.Get("/debug/pprof/heap", pprof.Handler("heap").ServeHTTP)
		m.Get("/debug/pprof/goroutine", pprof.Handler("goroutine").ServeHTTP)
		m.Get("/debug/pprof/threadcreate", pprof.Handler("threadcreate").ServeHTTP)
	}

	m.RunOnAddr(":" + *port)
}
