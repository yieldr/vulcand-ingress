package oauth2

import (
	"net/http"
	"net/url"

	"github.com/codegangsta/cli"
	"github.com/gorilla/context"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/vulcand/vulcand/plugin"
	"golang.org/x/oauth2"
)

var SessionStore sessions.Store

func init() {
	SessionStore = sessions.NewCookieStore(securecookie.GenerateRandomKey(32))
}

func GetSpec() *plugin.MiddlewareSpec {
	return &plugin.MiddlewareSpec{
		Type:      "oauth2",
		FromOther: FromOther,
		FromCli:   FromCli,
		CliFlags:  CliFlags(),
	}
}

func FromOther(o OAuth2) (plugin.Middleware, error) {
	return New(o.Domain, o.ClientID, o.ClientSecret, o.RedirectURL), nil
}

func FromCli(c *cli.Context) (plugin.Middleware, error) {
	return New(
		c.String("domain"),
		c.String("clientId"),
		c.String("clientSecret"),
		c.String("redirectUrl")), nil
}

func CliFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{Name: "domain", Usage: "oauth2 idp domain"},
		cli.StringFlag{Name: "clientId", Usage: "oauth2 client id"},
		cli.StringFlag{Name: "clientSecret", Usage: "oauth2 client secret"},
		cli.StringFlag{Name: "redirectUrl", Usage: "oauth2 redirect url"},
	}
}

type OAuth2 struct {
	Domain string
	*oauth2.Config
}

func New(domain, clientID, clientSecret, redirectURL string) *OAuth2 {
	return &OAuth2{
		Domain: domain,
		Config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Scopes:       []string{"openid", "profile"},
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://" + domain + "/authorize",
				TokenURL: "https://" + domain + "/oauth/token",
			},
		},
	}
}

func (o *OAuth2) NewHandler(next http.Handler) (http.Handler, error) {
	return context.ClearHandler(&OAuth2Handler{
		conf: o,
		next: next,
	}), nil
}

type OAuth2Handler struct {
	conf *OAuth2
	next http.Handler
}

func (o *OAuth2Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/callback" {
		o.Callback(w, r)
	} else {
		o.All(w, r)
	}
}

func (o *OAuth2Handler) Callback(w http.ResponseWriter, r *http.Request) {

	t, err := o.conf.Exchange(oauth2.NoContext, r.URL.Query().Get("code"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s, err := SessionStore.Get(r, "oauth2-session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !t.Valid() {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	s.Values["access_token"] = t.AccessToken
	s.Values["id_token"] = t.Extra("id_token")

	if err = s.Save(r, w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func (o *OAuth2Handler) All(w http.ResponseWriter, r *http.Request) {

	s, err := SessionStore.Get(r, "oauth2-session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if _, ok := s.Values["access_token"]; !ok {

		q := make(url.Values)
		q.Set("client", o.conf.ClientID)
		q.Set("redirect_uri", o.conf.RedirectURL)
		q.Set("protocol", "oauth2")
		q.Set("response_type", "code")

		http.Redirect(w, r, "https://"+o.conf.Domain+"/login?"+q.Encode(), 302)
		return
	}

	o.next.ServeHTTP(w, r)
}
