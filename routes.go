package main

import (
	"net/http"

	"github.com/Xe/hideyhole-site/database"
	"github.com/Xe/hideyhole-site/discordwidget"
	"github.com/Xe/hideyhole-site/interop"
	"github.com/Xe/martini-oauth2"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/sessions"
	acerender "github.com/yosssi/martini-acerender"
)

type Wrapper struct {
	Data    interface{}
	Session sessions.Session
}

func (si *Site) doError(w http.ResponseWriter, req *http.Request, code int, why string) {
	si.log.Printf("%s %s %s %s: %v", req.RemoteAddr, req.Method, req.RequestURI, req.RemoteAddr, why)
	http.Error(w, why, code)
}

func (si *Site) renderTemplate(code int, templateName string, data interface{}, s sessions.Session, r acerender.Render) {
	r.HTML(code, "base:"+templateName, Wrapper{
		Session: s,
		Data:    data,
	}, nil)
}

func (si *Site) getIndex(w http.ResponseWriter, req *http.Request, s sessions.Session, r acerender.Render) {
	guild, err := discordwidget.GetGuild(*guildID)

	if err != nil {
		si.doError(w, req, http.StatusInternalServerError, "Couldn't get guild information")
		return
	}

	si.renderTemplate(http.StatusOK, "index", guild, s, r)
}

func (si *Site) getChat(w http.ResponseWriter, req *http.Request, s sessions.Session, r acerender.Render) {
	guild, err := discordwidget.GetGuild(*guildID)

	if err != nil {
		si.doError(w, req, http.StatusInternalServerError, "Couldn't get guild information")
		return
	}

	si.renderTemplate(http.StatusOK, "chat", guild, s, r)
}

func (si *Site) getMyProfile(w http.ResponseWriter, req *http.Request, s sessions.Session, t moauth2.Tokens, r acerender.Render) {
	dUser, err := interop.GetOwnDiscordUser(t)
	if err != nil {
		si.doError(w, req, http.StatusInternalServerError, err.Error())
		return
	}

	data := struct {
		User *interop.DiscordUser
	}{
		User: dUser,
	}

	si.renderTemplate(http.StatusOK, "profile", data, s, r)
}

func (si *Site) getHealth() (int, string) {
	return 200, "okay"
}

func (si *Site) getUserByID(w http.ResponseWriter, req *http.Request, s sessions.Session, r acerender.Render, params martini.Params) {
	user, err := si.db.GetUser(params["id"])

	if err != nil {
		switch err {
		case database.ErrNoUserFound:
			si.doError(w, req, http.StatusNotFound, "No such user exists")
			return
		default:
			si.doError(w, req, http.StatusInternalServerError, err.Error())
			return
		}
	}

	data := struct {
		User *interop.DiscordUser
	}{
		User: user,
	}

	si.renderTemplate(http.StatusOK, "profile", data, s, r)
}
