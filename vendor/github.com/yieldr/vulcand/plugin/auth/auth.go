package auth

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/alexkappa/errors"
	"github.com/codegangsta/cli"
	"github.com/mailgun/log"
	"github.com/vulcand/vulcand/plugin"
	"github.com/yieldr/vulcand/pkg/cache"
)

func GetLegacySpec() *plugin.MiddlewareSpec {
	return &plugin.MiddlewareSpec{
		Type:      "yieldrauth",
		FromOther: FromOther,
		FromCli:   FromCli,
		CliFlags:  CliFlags(),
	}
}

func GetSpec() *plugin.MiddlewareSpec {
	return &plugin.MiddlewareSpec{
		Type:      "auth",
		FromOther: FromOther,
		FromCli:   FromCli,
		CliFlags:  CliFlags(),
	}
}

func CliFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:  "server-url",
			Usage: "Authentication server URL",
			Value: "https://crowd.ydworld.com",
		},
		cli.StringFlag{
			Name:  "username",
			Usage: "Authentication server username",
			Value: "",
		},
		cli.StringFlag{
			Name:  "password",
			Usage: "Authentication server password",
			Value: "",
		},
		cli.DurationFlag{
			Name:  "cache-expiration",
			Usage: "Authentication cache expiration",
			Value: 1 * time.Minute,
		},
	}
}

// FromOther constructs the middleware from another middleware struct, typically
// originating from a JSON object.
func FromOther(o AuthMiddleware) (plugin.Middleware, error) {
	return NewAuthMiddleware(o.ServerURL, o.Username, o.Password, o.CacheExpiration), nil
}

// FromCli constructs the middleware from the command line
func FromCli(c *cli.Context) (plugin.Middleware, error) {
	return NewAuthMiddleware(
		c.String("server-url"),
		c.String("username"),
		c.String("password"),
		c.Duration("cache-expiration"),
	), nil
}

type AuthMiddleware struct {
	// Authentication server URL
	ServerURL string
	// Username to use when interacting with the authentication server.
	Username string
	// Password to use when interacting with the authentication server.
	Password string
	// Duration to keep users cached.
	CacheExpiration time.Duration
}

// NewAuthMiddleware creates the authentication middleware.
func NewAuthMiddleware(url, username, password string, expiry time.Duration) *AuthMiddleware {
	return &AuthMiddleware{url, username, password, expiry}
}

// NewHandler is required by vulcand to register itself as frontend middleware.
func (a *AuthMiddleware) NewHandler(next http.Handler) (http.Handler, error) {
	return NewAuthHandler(a, next), nil
}

// String is a helper method used by vulcand log.
func (a *AuthMiddleware) String() string {
	return fmt.Sprintf(
		"auth-server-url=%s, auth-username=%s auth-password=%s auth-cache-expiration=%s",
		a.ServerURL,
		a.Username,
		Mask(a.Password, '*'),
		a.CacheExpiration,
	)
}

// AuthHandler is a http.Handler that executes the auth middleware.
type AuthHandler struct {
	provider Provider
	cache    cache.Cache
	next     http.Handler
}

// ServeHTTP satifies the http.Handler interface.
func (a *AuthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	u, err := a.authenticate(r.BasicAuth())
	if err != nil {
		Unauthorized(w, err)
		return
	}
	r.Header.Set("X-Auth-Username", u.Username())
	r.Header.Set("X-Auth-FullName", u.FullName())
	r.Header.Set("X-Auth-Email", u.Email())
	r.Header.Set("X-Auth-Accounts", strings.Join(u.Accounts(), ","))
	r.Header.Set("X-Auth-Roles", strings.Join(u.Roles(), ","))
	a.next.ServeHTTP(w, r)
}

func (a *AuthHandler) authenticate(username, password string, ok bool) (user User, err error) {
	if !ok {
		return nil, errors.New("No username or password given.")
	}

	key := Base64(fmt.Sprintf("%s:%s", username, password))

	if value, ok := a.cache.Get(key); ok {
		log.Debugf("auth: user %s found in cache.", username)
		a.cache.Set(key, value, cache.DefaultExpiration)
		return value.(User), nil
	}

	log.Debugf("auth: authenticating user %s with crowd.", username)
	user, err = a.provider.Authenticate(username, password)
	if err != nil {
		return
	}
	a.cache.Set(key, user, cache.DefaultExpiration)

	return user, nil
}

// NewAuthHandler creates a new auth handler with the given configuration and
// the next handler in the chain.
func NewAuthHandler(a *AuthMiddleware, next http.Handler) *AuthHandler {
	return &AuthHandler{
		provider: NewCrowdProvider(a.ServerURL, a.Username, a.Password),
		cache:    cache.NewCache(a.CacheExpiration),
		next:     next,
	}
}

// Unauthorized is a helper HTTP response that sets the status code to 401 and
// sets the WWW-Authenticate accordingly.
func Unauthorized(w http.ResponseWriter, err error) {
	w.Header().Set("WWW-Authenticate", `Basic realm="my.yieldr.com"`)
	Error(w, err, http.StatusUnauthorized)
}

// Error is a helper HTTP response that marshals the error and writes it to the
// underlying http.ResponseWriter
func Error(w http.ResponseWriter, err error, code int) {
	b, _ := json.Marshal(err)
	if e, ok := err.(errors.Error); ok {
		w.Header().Set("X-Error", e.Message())
	}
	w.WriteHeader(code)
	w.Write(b)
}

// Mask masks a string to conceal sensitive information. The masked string will
// have the same length as the original.
func Mask(s string, mask rune) string {
	m := make([]rune, len(s))
	for i, _ := range s {
		m[i] = mask
	}
	return string(m)
}

// Base64 encodes a string as Base 64 and returns it.
func Base64(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}
