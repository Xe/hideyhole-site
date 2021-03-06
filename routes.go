package main

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/Xe/hideyhole-site/database"
	"github.com/Xe/hideyhole-site/discordwidget"
	"github.com/Xe/hideyhole-site/interop"
	"github.com/Xe/martini-oauth2"
	"github.com/extemporalgenome/slug"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/sessions"
	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday"
	acerender "github.com/yosssi/martini-acerender"
)

type Wrapper struct {
	Data    interface{}
	Session sessions.Session
}

func (si *Site) doError(w http.ResponseWriter, req *http.Request, code int, why string) {
	si.log.Printf(
		"[%s] %s %s %d: %v",
		req.RemoteAddr, req.Method, req.RequestURI, code, why,
	)
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

const (
	ficsPerPage = 10
)

func (si *Site) listFics(w http.ResponseWriter, req *http.Request, s sessions.Session, r acerender.Render, params martini.Params) {
	if params["page"] == "" {
		params["page"] = "1"
	}

	pageNum, err := strconv.Atoi(params["page"])
	if err != nil {
		si.doError(w, req, http.StatusBadRequest, "invalid page number \""+params["page"]+"\"")
		return
	}

	fics, err := si.db.GetFics(ficsPerPage, pageNum-1)
	if err != nil {
		log.Println(err)
		si.doError(w, req, http.StatusInternalServerError, "cannot fetch fics")
		return
	}

	data := struct {
		Pagenum int
		Fics    []database.Fic
		Route   string
	}{
		Pagenum: pageNum,
		Fics:    fics,
		Route:   "fics/index",
	}

	si.renderTemplate(http.StatusOK, "ficlist", data, s, r)
}

func (si *Site) getCreateFic(s sessions.Session, r acerender.Render) {
	si.renderTemplate(http.StatusOK, "ficcreate", nil, s, r)
}

type CreateFicForm struct {
	Name        string `form:"name" binding:"required"`
	Description string `form:"description" binding:"required"`
}

func (si *Site) postCreateFic(w http.ResponseWriter, req *http.Request, form CreateFicForm, s sessions.Session) {
	id := si.db.IDGen.Generate()

	unsafe := blackfriday.MarkdownCommon([]byte(form.Description))
	html := bluemonday.UGCPolicy().SanitizeBytes(unsafe)

	fic := &database.Fic{
		ID:       id.String(),
		Created:  time.Now(),
		AuthorID: s.Get("uid").(string),
		Title:    form.Name,

		Description:     form.Description,
		DescriptionHTML: string(html),
	}

	err := si.db.PutFic(fic)
	if err != nil {
		si.log.Println(err)
		si.doError(w, req, http.StatusInternalServerError, "data not saved, please try again")
		return
	}

	http.Redirect(w, req, "/fics/"+slug.Slug(fic.Title)+"/"+fic.ID, http.StatusFound)
}

func (si *Site) getFic(w http.ResponseWriter, req *http.Request, s sessions.Session, r acerender.Render, params martini.Params) {
	id := params["id"]
	log.Println(id)

	fic, chapters, err := si.db.GetFicAndChapters(id)
	if err != nil {
		switch err {
		case database.ErrNoFicFound:
			si.doError(w, req, http.StatusNotFound, "not found")
			return
		default:
			log.Println(err)
			si.doError(w, req, http.StatusInternalServerError, "cannot fetch fic")
			return
		}
	}

	data := struct {
		Fic      *database.Fic
		Chapters map[string]database.Chapter
	}{
		Fic:      fic,
		Chapters: chapters,
	}

	si.renderTemplate(http.StatusOK, "fic", data, s, r)
}
