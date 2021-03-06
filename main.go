package main

import (
	"context"
	"errors"
	"flag"
	"html/template"
	"log"
	"net/http"
	"net/http/pprof"

	"cloud.google.com/go/preview/logging"
	"github.com/Xe/hideyhole-site/database"
	"github.com/Xe/hideyhole-site/interop"
	"github.com/Xe/hideyhole-site/oauth2/discord"
	"github.com/Xe/martini-oauth2"
	"github.com/extemporalgenome/slug"
	"github.com/facebookgo/flagconfig"
	"github.com/facebookgo/flagenv"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/binding"
	"github.com/martini-contrib/csrf"
	"github.com/martini-contrib/sessions"
	"github.com/yosssi/ace"
	"github.com/yosssi/martini-acerender"
	"golang.org/x/oauth2"
)

var (
	clientID                 = flag.String("discord-client-id", "", "discord oauth client id")
	clientSecret             = flag.String("discord-client-secret", "", "discord oauth client secret")
	googleProjectID          = flag.String("google-project-id", "", "google project ID")
	googleDatastoreNamespace = flag.String("google-datastore-namespace", "hideyhole-other", "google datastore namespace")
	port                     = flag.String("port", "3093", "TCP port to listen on for HTTP requests")
	guildID                  = flag.String("guild-id", "", "guild ID for allowing membership")
	cookieKey                = flag.String("cookie-key", "", "random cookie key")
	salt                     = flag.String("salt", "", "salt for any passwords or crypto stuff")
	debug                    = flag.Bool("debug", false, "add /debug routes? pprof, etc.")
	domain                   = flag.String("domain", "localhost:3093", "redirect URL base for OAuth")

	discordOAuthClient *oauth2.Config
)

type Site struct {
	db        *database.Database
	logClient *logging.Client
	log       *log.Logger
}

func (si *Site) populateInfo(s sessions.Session, t moauth2.Tokens) error {
	otoken := s.Get("oauth2_token")
	if otoken == nil {
		return nil
	}

	uid := s.Get("uid")
	if uid == nil {
		dUser, err := interop.GetOwnDiscordUser(t)
		if err != nil {
			return err
		}

		s.Set("uid", dUser.ID)
		s.Set("username", dUser.Username)
		s.Set("avatarhash", dUser.Avatar)

		err = si.db.PutUser(dUser)
		if err != nil {
			return err
		}

		guilds, err := interop.GetOwnDiscordGuilds(t)
		if err != nil {
			return err
		}

		ok := false

		for _, guild := range guilds {
			if guild.ID == *guildID {
				ok = true

				break
			}
		}

		if !ok {
			return errors.New("Not in target guild")
		}

		si.log.Printf("user login: %s %q", dUser.ID, dUser.Username)
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
		RedirectURL:  *domain + moauth2.PathCallback,
	}

	logClient, err := logging.NewClient(context.Background(), *googleProjectID)
	if err != nil {
		log.Fatal(err)
	}

	logger := logClient.Logger(
		*googleDatastoreNamespace+"_"+martini.Env,
		logging.CommonLabels(map[string]string{
			"namespace": *googleDatastoreNamespace,
			"env":       martini.Env,
		}),
	).StandardLogger(
		logging.Default,
	)

	db, err := database.Init(*googleDatastoreNamespace, *googleProjectID)
	if err != nil {
		log.Fatal(err)
	}

	si := &Site{
		db:        db,
		logClient: logClient,
		log:       logger,
	}

	m := martini.Classic()
	store := sessions.NewCookieStore([]byte(*cookieKey))

	m.Use(sessions.Sessions("backplane.cadeyforum", store))
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
				"slug": slug.Slug,
				"dec": func(a int) int {
					return a - 1
				},
				"inc": func(a int) int {
					return a + 1
				},
				"sget": func(s sessions.Session, key string) string {
					val := s.Get(key)
					if val == nil {
						return ""
					}

					return val.(string)
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

	// Routes
	m.Get("/", si.getIndex)
	m.Get("/chat", si.getChat)
	m.Get("/health", si.getHealth)

	m.Group("/profile", func(r martini.Router) {
		r.Get("/me", si.getMyProfile)
		r.Get("/:slug/:id", si.getUserByID)
		r.Get("/:id", si.getUserByID)
	}, moauth2.LoginRequired)

	m.Group("/fics", func(r martini.Router) {
		r.Get("", si.listFics)
		r.Get("/", si.listFics)
		r.Get("/index/:page", si.listFics)
		r.Get("/create", si.getCreateFic)
		r.Post("/create", binding.Bind(CreateFicForm{}), si.postCreateFic)
		r.Get("/:slug/:id", si.getFic)
	}, moauth2.LoginRequired)

	if *debug {
		log.Printf("Adding /debug routes")
		if martini.Env == martini.Prod {
			log.Printf("The pprof routes are enabled in production!!! Please act with care.")
		}

		m.Group("/debug/pprof", func(r martini.Router) {
			r.Get("", pprof.Index)
			r.Get("/cmdline", pprof.Cmdline)
			r.Get("/profile", pprof.Profile)
			r.Get("/symbol", pprof.Symbol)
			r.Post("/symbol", pprof.Symbol)
			r.Get("/block", pprof.Handler("block").ServeHTTP)
			r.Get("/heap", pprof.Handler("heap").ServeHTTP)
			r.Get("/goroutine", pprof.Handler("goroutine").ServeHTTP)
			r.Get("/threadcreate", pprof.Handler("threadcreate").ServeHTTP)
		})
	}

	m.RunOnAddr(":" + *port)
}
