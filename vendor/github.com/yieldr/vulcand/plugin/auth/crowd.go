package auth

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

const (
	authURI = "/rest/usermanagement/latest/authentication"
	userURI = "/rest/usermanagement/latest/user"
)

type Provider interface {
	Authenticate(username, password string) (User, error)
}

type Crowd struct {
	serverURL  string
	username   string
	password   string
	httpClient *http.Client
}

// Authenticate checks the users credentials and returns the user details upon
// success.
func (c *Crowd) Authenticate(username, password string) (User, error) {
	if err := c.checkPassword(username, password); err != nil {
		return nil, err
	}
	return c.getUser(username)
}

// NewCrowdProvider creates a new Crowd auth provider.
func NewCrowdProvider(url, username, password string) Provider {
	// TODO: Fix the certificate on the server to avoid this measure.
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	return &Crowd{
		serverURL:  url,
		username:   username,
		password:   password,
		httpClient: &http.Client{Transport: transport},
	}
}

func (c *Crowd) checkPassword(username, password string) error {
	url := c.authRequestURL(username)
	body := c.authRequestBody(password)

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return err
	}
	req.SetBasicAuth(c.username, c.password)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	res, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		return errors.New("Invalid username or password.")
	}

	return nil
}

func (c *Crowd) getUser(username string) (*crowdUser, error) {
	url := c.userRequestURL(username)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(c.username, c.password)
	req.Header.Set("Accept", "application/json")

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	u := new(crowdUser)
	err = json.Unmarshal(b, u)
	return u, err
}

func (c *Crowd) authRequestURL(username string) string {
	return fmt.Sprintf("%s%s?username=%s", c.serverURL, authURI, username)
}

func (c *Crowd) authRequestBody(password string) io.Reader {
	body := bytes.NewBuffer(nil)
	body.WriteString(fmt.Sprintf(`{"value": "%s"}`, password))
	return body
}

func (c *Crowd) userRequestURL(username string) string {
	return fmt.Sprintf("%s%s?expand=attributes&username=%s", c.serverURL, userURI, username)
}

type crowdUser struct {
	Name        string `json:"name"`
	DisplayName string `json:"display-name"`
	EMail       string `json:"email"`
	Key         string `json:"key"`
	Attributes  struct {
		Attributes []struct {
			Name   string   `json:"name"`
			Values []string `json:"values"`
		} `json:"attributes"`
	} `json:"attributes"`
}

func (u crowdUser) Attribute(name string) []string {
	for _, attr := range u.Attributes.Attributes {
		if attr.Name == name {
			return attr.Values
		}
	}
	return nil
}

func (u crowdUser) Username() string { return u.Name }
func (u crowdUser) FullName() string { return u.DisplayName }
func (u crowdUser) Email() string    { return u.EMail }

func (u crowdUser) Accounts() []string {
	return u.Attribute("companyId")
}

func (u crowdUser) Roles() []string {
	return u.Attribute("roles")
}
