// Package authproxy implements a connector which relies on external
// authentication (e.g. mod_auth in Apache2) and returns an identity with the
// HTTP header X-Remote-User as verified email.
package authproxy

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/dexidp/dex/connector"
	"github.com/dexidp/dex/pkg/log"
)

// Config holds the configuration parameters for a connector which returns an
// identity with the HTTP header X-Remote-User as verified email.
type Config struct {
	HeaderName string   `json:"headerName"`
	Groups     []string `json:"groups"`
}

// Open returns an authentication strategy which requires no user interaction.
func (c *Config) Open(id string, logger log.Logger) (connector.Connector, error) {
	if c.HeaderName == "" {
		c.HeaderName = "X-Remote-User"
	}
	return &callback{headerName: c.HeaderName, logger: logger, pathSuffix: "/" + id, groups: c.Groups}, nil
}

// Callback is a connector which returns an identity with the HTTP header
// X-Remote-User as verified email.
type callback struct {
	headerName string
	groups     []string
	logger     log.Logger
	pathSuffix string
}

// LoginURL returns the URL to redirect the user to login with.
func (m *callback) LoginURL(s connector.Scopes, callbackURL, state string) (string, error) {
	u, err := url.Parse(callbackURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse callbackURL %q: %v", callbackURL, err)
	}
	u.Path += m.pathSuffix
	v := u.Query()
	v.Set("state", state)
	u.RawQuery = v.Encode()
	return u.String(), nil
}

// HandleCallback parses the request and returns the user's identity
func (m *callback) HandleCallback(s connector.Scopes, r *http.Request) (connector.Identity, error) {
	remoteUser := r.Header.Get(m.headerName)
	if remoteUser == "" {
		return connector.Identity{}, fmt.Errorf("required HTTP header %s is not set", m.headerName)
	}
	// TODO: add support for X-Remote-Group, see
	// https://kubernetes.io/docs/admin/authentication/#authenticating-proxy
	return connector.Identity{
		UserID:        remoteUser, // TODO: figure out if this is a bad ID value.
		Email:         remoteUser,
		EmailVerified: true,
		Groups:        m.groups,
	}, nil
}
