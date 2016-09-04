package main

import (
	"log"
	"net/http"

	"github.com/Xe/hideyhole-site/discordwidget"
	"github.com/martini-contrib/oauth2"
	"github.com/martini-contrib/sessions"
	acerender "github.com/yosssi/martini-acerender"
)

func (si *Site) doError(w http.ResponseWriter, req *http.Request, code int, why string) {
	log.Printf("%s %s %s: %v", req.Method, req.RequestURI, req.RemoteAddr, why)
	http.Error(w, why, code)
}

func (si *Site) renderTemplate(code int, templateName string, data interface{}, s sessions.Session, r acerender.Render) {
	r.HTML(code, "base:"+templateName, Wrapper{
		Session: s,
		Data:    data,
	}, nil)
}

func (si *Site) getIndex(s sessions.Session, r acerender.Render) {
	si.renderTemplate(http.StatusOK, "test", nil, s, r)
}

func (si *Site) getChat(w http.ResponseWriter, req *http.Request, s sessions.Session, r acerender.Render) {
	guild, err := discordwidget.GetGuild(*guildID)

	if err != nil {
		si.doError(w, req, http.StatusInternalServerError, "Couldn't get guild information")
		return
	}

	si.renderTemplate(http.StatusOK, "chat", guild, s, r)
}

func (si *Site) logout(s sessions.Session, w http.ResponseWriter, r *http.Request) {
	s.Clear()

	http.Redirect(w, r, "/", http.StatusFound)
}

func (si *Site) getMyProfile(w http.ResponseWriter, req *http.Request, s sessions.Session, t oauth2.Tokens, r acerender.Render) {
	dUser, err := si.getOwnDiscordUser(t)
	if err != nil {
		si.doError(w, req, http.StatusInternalServerError, err.Error())
		return
	}

	data := struct {
		User DiscordUser
	}{
		User: *dUser,
	}

	log.Printf("%#v", data)

	si.renderTemplate(http.StatusOK, "profile", data, s, r)
}
