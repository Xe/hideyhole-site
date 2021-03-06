// Copyright 2014 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package moauth2 contains Martini handlers to provide
// user login via an OAuth 2.0 backend.
package moauth2

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/go-martini/martini"
	"github.com/martini-contrib/sessions"
	"golang.org/x/oauth2"
)

const (
	codeRedirect = 302
	keyToken     = "oauth2_token"
	keyNextPage  = "next"
)

var (
	// PathLogin is the path to handle OAuth 2.0 logins.
	PathLogin = "/login"
	// PathLogout is the path to handle OAuth 2.0 logouts.
	PathLogout = "/logout"
	// PathCallback is the path to handle callback from OAuth 2.0 backend
	// to exchange credentials.
	PathCallback = "/oauth2callback"
	// PathError is the path to handle error cases.
	PathError = "/oauth2error"
)

// Tokens represents a container that contains user's OAuth 2.0 access and refresh tokens.
type Tokens interface {
	Access() string
	Refresh() string
	Expired() bool
	ExpiryTime() time.Time
}

type token struct {
	oauth2.Token
}

// Access returns the access token.
func (t *token) Access() string {
	return t.AccessToken
}

// Refresh returns the refresh token.
func (t *token) Refresh() string {
	return t.RefreshToken
}

// Expired returns whether the access token is expired or not.
func (t *token) Expired() bool {
	if t == nil {
		return true
	}
	return !t.Token.Valid()
}

// ExpiryTime returns the expiry time of the user's access token.
func (t *token) ExpiryTime() time.Time {
	return t.Expiry
}

// String returns the string representation of the token.
func (t *token) String() string {
	return fmt.Sprintf("tokens: %v", t.String())
}

// NewOAuth2Provider returns a generic OAuth 2.0 backend endpoint.
func NewOAuth2Provider(conf *oauth2.Config, then func(sessions.Session, Tokens) error) martini.Handler {
	return func(s sessions.Session, c martini.Context, w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			switch r.URL.Path {
			case PathLogin:
				login(conf, s, w, r)
			case PathLogout:
				logout(s, w, r)
			case PathCallback:
				handleOAuth2Callback(conf, s, w, r, then)
			}
		}
		tk := unmarshallToken(s)
		if tk != nil {
			// check if the access token is expired
			if tk.Expired() && tk.Refresh() == "" {
				s.Delete(keyToken)
				tk = nil
			}
		}
		// Inject tokens.
		c.MapTo(tk, (*Tokens)(nil))
	}
}

// Handler that redirects user to the login page
// if user is not logged in.
// Sample usage:
// m.Get("/login-required", oauth2.LoginRequired, func() ... {})
var LoginRequired = func() martini.Handler {
	return func(s sessions.Session, c martini.Context, w http.ResponseWriter, r *http.Request) {
		token := unmarshallToken(s)
		if token == nil || token.Expired() {
			next := url.QueryEscape(r.URL.RequestURI())
			http.Redirect(w, r, PathLogin+"?next="+next, codeRedirect)
		}
	}
}()

func login(f *oauth2.Config, s sessions.Session, w http.ResponseWriter, r *http.Request) {
	next := extractPath(r.URL.Query().Get(keyNextPage))
	if s.Get(keyToken) == nil {
		// User is not logged in.
		if next == "" {
			next = "/"
		}
		http.Redirect(w, r, f.AuthCodeURL(next), codeRedirect)
		return
	}
	// No need to login, redirect to the next page.
	http.Redirect(w, r, next, codeRedirect)
}

func logout(s sessions.Session, w http.ResponseWriter, r *http.Request) {
	next := extractPath(r.URL.Query().Get(keyNextPage))
	s.Clear()
	http.Redirect(w, r, next, codeRedirect)
}

func handleOAuth2Callback(f *oauth2.Config, s sessions.Session, w http.ResponseWriter, r *http.Request, then func(sessions.Session, Tokens) error) {
	next := extractPath(r.URL.Query().Get("state"))
	code := r.URL.Query().Get("code")
	t, err := f.Exchange(oauth2.NoContext, code)
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, PathError, codeRedirect)
		return
	}
	// Store the credentials in the session.
	val, _ := json.Marshal(t)
	s.Set(keyToken, val)

	tk := unmarshallToken(s)

	err = then(s, tk)
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, PathError, codeRedirect)
		return
	}

	http.Redirect(w, r, next, codeRedirect)
}

func unmarshallToken(s sessions.Session) (t *token) {
	if s.Get(keyToken) == nil {
		return
	}
	data := s.Get(keyToken).([]byte)
	var tk oauth2.Token
	json.Unmarshal(data, &tk)
	return &token{tk}
}

func extractPath(next string) string {
	n, err := url.Parse(next)
	if err != nil {
		return "/"
	}
	return n.Path
}
