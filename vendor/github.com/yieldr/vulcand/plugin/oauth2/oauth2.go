package oauth2

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"

	"github.com/alexkappa/errors"
	"github.com/codegangsta/cli"
	"github.com/gorilla/context"
	"github.com/gorilla/handlers"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/sirupsen/logrus"
	"github.com/vulcand/vulcand/plugin"
	"golang.org/x/oauth2"
)

const (
	errInvalidRedirect  = "invalid redirect uri"
	errInvalidState     = "invalid state parameter"
	errInvalidToken     = "invalid token"
	errFailWriteSession = "failed writing session cookie"
	errFailReadSession  = "failed reading session cookie"
	errFailExchangeCode = "failed exchanging oauth2 code for token"
)

var (
	// We will store cookies in the users browser to maintain an authenticated
	// session.
	sessionStore sessions.Store

	// Define a logger with some default fields used throughout the plugin.
	logger *logrus.Entry
)

func init() {
	sessionStore = sessions.NewCookieStore(securecookie.GenerateRandomKey(32))
	logger = logrus.WithField("plugin", "oauth2")
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
	logger.WithFields(logrus.Fields{
		"issuer_url":    o.IssuerURL,
		"client_id":     o.ClientID,
		"client_secret": o.ClientSecret,
		"redirect_url":  o.RedirectURL,
	}).Debugf("initializing from json")
	return New(o.IssuerURL, o.ClientID, o.ClientSecret, o.RedirectURL)
}

func FromCli(c *cli.Context) (plugin.Middleware, error) {
	return New(
		c.String("issuerUrl"),
		c.String("clientId"),
		c.String("clientSecret"),
		c.String("redirectUrl"))
}

func CliFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{Name: "issuerUrl", Usage: "oauth2 idp issuer url"},
		cli.StringFlag{Name: "clientId", Usage: "oauth2 client id"},
		cli.StringFlag{Name: "clientSecret", Usage: "oauth2 client secret"},
		cli.StringFlag{Name: "redirectUrl", Usage: "oauth2 redirect url"},
	}
}

type OAuth2 struct {
	// IssuerURL holds the authorization servers endpoint. This can be any OAuth2
	// compatible server however this plugin has only been tested to work with
	// Auth0.
	IssuerURL string
	// RedirectURLPath holds the URL path of the redirect URL. This path will be
	// reserved for handling the OAuth2 redirect callback therefore it is
	// important to use a path that does not conflict with upstream services
	// routing.
	RedirectURLPath string

	*oauth2.Config
}

// New creates a new OAuth2 plugin which can be used with vulcand. You can find
// the arguments to use at your preferred Identity Provider (IdP).
//
// This plugin is tested to work with Auth0.
func New(issuerURL, clientID, clientSecret, redirectURL string) (*OAuth2, error) {
	// The redirectURL should be stripped from anything that isnt a URL path. So
	// we parse it and ignore the rest.
	u, err := url.Parse(redirectURL)
	if err != nil {
		return nil, errors.Wrap(err, errInvalidRedirect)
	}
	// Now configure OAuth2 with the supplied information. Keep in mind that
	// OAuth2.Config.RedirectURL is not set right now as it will be decided at
	// run time by OAuth2.RedirectURL plus the request's Hostname if the path is
	// relative.
	return &OAuth2{
		IssuerURL:       issuerURL,
		RedirectURLPath: u.Path,
		Config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Scopes:       []string{"openid", "profile"},
			Endpoint: oauth2.Endpoint{
				AuthURL:  issuerURL + "/authorize",
				TokenURL: issuerURL + "/oauth/token",
			},
		},
	}, nil
}

// NewHandler returns the http.Handler which executes the OAuth2 plugin. It is
// wrapped in a recovery handler and a handler that clears the context from each
// request.
func (o *OAuth2) NewHandler(next http.Handler) (http.Handler, error) {
	var h http.Handler
	h = &OAuth2Handler{oauth2: o, next: next}
	h = context.ClearHandler(h)
	h = handlers.RecoveryHandler(handlers.RecoveryLogger(logrus.StandardLogger()))(h)
	return h, nil
}

func (o *OAuth2) String() string {
	return fmt.Sprintf(
		"issuer-url=%s, client-id=%s client-secret=%s redirect-url=%s",
		o.IssuerURL,
		o.ClientID,
		o.ClientSecret,
		o.RedirectURL,
	)
}

type OAuth2Handler struct {
	oauth2 *OAuth2
	next   http.Handler
}

func (h *OAuth2Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == h.oauth2.RedirectURLPath {
		h.Callback(w, r)
	} else {
		h.All(w, r)
	}
}

func (h *OAuth2Handler) Callback(w http.ResponseWriter, r *http.Request) {

	s, err := sessionStore.Get(r, "oauth2-session")
	if err != nil {
		logger.WithField("error", err).Errorf(errFailReadSession)
		http.Error(w, errors.Wrap(err, errFailReadSession).Error(), http.StatusInternalServerError)
		return
	}

	// Check the state parameter. This should match the state we set during the
	// authentication redirect.
	if _, ok := s.Values["state"]; !ok || s.Values["state"].(string) != r.URL.Query().Get("state") {
		logger.WithField("error", err).Errorf(errInvalidState)
		http.Error(w, errInvalidState, http.StatusBadRequest)
		return
	}

	// Exchange the oauth2 authentication code with a token.
	t, err := h.oauth2.Exchange(oauth2.NoContext, r.URL.Query().Get("code"))
	if err != nil {
		logger.WithField("error", err).Errorf(errFailExchangeCode)
		http.Error(w, errors.Wrap(err, errFailExchangeCode).Error(), http.StatusInternalServerError)
		return
	}
	if !t.Valid() {
		logger.WithField("error", err).Errorf(errInvalidToken)
		http.Error(w, errInvalidToken, http.StatusUnauthorized)
		return
	}

	// Save access_token and id_token to the encrypted cookie. We'll use these
	// to authenticate the user in subsequent calls.
	s.Values["access_token"] = t.AccessToken

	if err = s.Save(r, w); err != nil {
		logger.WithField("error", err).Errorf(errFailWriteSession)
		http.Error(w, errors.Wrap(err, errFailWriteSession).Error(), http.StatusInternalServerError)
		return
	}

	// Redirect the user to the URL they intended to visit originally before
	// we redirected them to authenticate.
	returnTo := "/"
	if _, ok := s.Values["return_to"]; ok {
		returnTo = s.Values["return_to"].(string)
	}

	http.Redirect(w, r, returnTo, http.StatusTemporaryRedirect)
}

func (h *OAuth2Handler) All(w http.ResponseWriter, r *http.Request) {

	s, err := sessionStore.Get(r, "oauth2-session")
	if err != nil {
		logger.WithField("error", err).Errorf(errFailReadSession)
		http.Error(w, errors.Wrap(err, errFailReadSession).Error(), http.StatusInternalServerError)
		return
	}

	// Check for a valid access_token. If one is not present the user will be
	// redirected to the IdP for authentication.
	if _, ok := s.Values["access_token"]; !ok {
		// Generate a random base64 encoded hash to append to the authentication
		// request. We'll make sure it matches the state parameter in the
		// callback handler.
		b := make([]byte, 32)
		rand.Read(b)
		state := base64.StdEncoding.EncodeToString(b)
		s.Values["state"] = state

		// Record the current URL path so we can redirect the user here after
		// they authenticated.
		s.Values["return_to"] = r.RequestURI

		if err = s.Save(r, w); err != nil {
			logger.WithField("error", err).Errorf(errFailWriteSession)
			http.Error(w, errors.Wrap(err, errFailWriteSession).Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, h.oauth2.AuthCodeURL(state), http.StatusTemporaryRedirect)

		return
	}

	h.next.ServeHTTP(w, r)
}
