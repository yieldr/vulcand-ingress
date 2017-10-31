package vulcan

import (
	"strings"
	"time"

	"k8s.io/api/extensions/v1beta1"

	"github.com/vulcand/vulcand/api"
	"github.com/vulcand/vulcand/engine"
	"github.com/yieldr/vulcand/registry"

	"github.com/yieldr/vulcand-ingress/pkg/kubernetes/annotations"
)

type Client struct {
	*api.Client
}

func New(addr string) *Client {
	r, _ := registry.GetRegistry()
	return &Client{api.NewClient(addr, r)}
}

func (c *Client) DeleteBackend(ns, name string) error {
	// As the backend's ID is made up from the ingress name and service name we
	// list all the vulcan backends and check which ones start with the
	// backend's ID.
	backends, err := c.Client.GetBackends()
	if err != nil {
		return err
	}
	// If the backend matches the namespace and ingress name in prefix then we
	// delete it.
	for _, backend := range backends {
		if strings.HasPrefix(backend.Id, strings.Join([]string{ns, name}, ".")) {
			err = c.Client.DeleteBackend(backend.Key())
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *Client) DeleteFrontend(ns, name string) error {
	frontends, err := c.Client.GetFrontends()
	if err != nil {
		return err
	}
	for _, frontend := range frontends {
		if strings.HasPrefix(frontend.Id, strings.Join([]string{ns, name}, ".")) {
			err = c.Client.DeleteFrontend(frontend.Key())
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *Client) DeleteServer(ns, name string) error {
	servers, err := c.Client.GetServers()
	if err != nil {
		return err
	}
	for _, server := range servers {
		if strings.HasPrefix(server.Id, strings.Join([]string{ns, name}, ".")) {
			err = c.Client.DeleteServer(engine.ServerKey{
				BackendKey: ,
				Id:
				})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *Client) SyncBackend(ingress *v1beta1.Ingress, backend *v1beta1.IngressBackend) error {
	return c.UpsertBackend(engine.Backend{
		Id:   CreateID(ingress, backend),
		Type: engine.HTTP,
		Settings: engine.HTTPBackendSettings{
			Timeouts: engine.HTTPBackendTimeouts{
				Read:         annotations.GetString(ingress, annotations.ReadTimeout),
				Dial:         annotations.GetString(ingress, annotations.DialTimeout),
				TLSHandshake: annotations.GetString(ingress, annotations.TLSHandshakeTimeout),
			},
			KeepAlive: engine.HTTPBackendKeepAlive{
				Period:              annotations.GetString(ingress, annotations.KeepAlive),
				MaxIdleConnsPerHost: annotations.GetInt(ingress, annotations.MaxIdleConnsPerHost),
			},
		},
	})
}

func (c *Client) SyncFrontend(ingress *v1beta1.Ingress, backend *v1beta1.IngressBackend, host, path string) error {
	return c.UpsertFrontend(engine.Frontend{
		Id:        CreateID(ingress, backend),
		BackendId: CreateID(ingress, backend),
		Type:      engine.HTTP,
		Route:     CreateRoute(host, path),
		Settings: &engine.HTTPFrontendSettings{
			Limits: engine.HTTPFrontendLimits{
				MaxBodyBytes:    int64(annotations.GetInt(ingress, annotations.MaxBodyBytes)),
				MaxMemBodyBytes: int64(annotations.GetInt(ingress, annotations.MaxMemBodyBytes)),
			},
			FailoverPredicate:  annotations.GetString(ingress, annotations.FailoverPredicate),
			Hostname:           annotations.GetString(ingress, annotations.Hostname),
			TrustForwardHeader: annotations.GetBool(ingress, annotations.TrustForwardHeader),
		},
	}, time.Duration(0))
}

func (c *Client) SyncServer(ingress *v1beta1.Ingress, backend *v1beta1.IngressBackend) error {
	key := engine.BackendKey{
		Id: CreateID(ingress, backend),
	}
	server := engine.Server{
		Id:  CreateID(ingress, backend),
		URL: CreateURL(backend),
	}
	return c.UpsertServer(key, server, time.Duration(0))
}

func (c *Client) SyncMiddleware(ingress *v1beta1.Ingress, backend *v1beta1.IngressBackend) error {
	for key, value := range annotations.GetMiddleware(ingress) {
		// Retrieve the middleware specification from the vulcand plugin
		// registry.
		spec := c.Registry.GetSpec(key)
		if spec != nil {
			// Parse the middleware configuration from a JSON payload.
			m, err := spec.FromJSON([]byte(value))
			if err != nil {
				return err
			}
			// Now upsert the middleware to the vulcand API.
			err = c.UpsertMiddleware(
				engine.FrontendKey{
					Id: backend.ServiceName,
				}, engine.Middleware{
					Id:         CreateID(ingress, backend, key),
					Type:       key,
					Middleware: m,
				},
				time.Duration(0))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func CreateID(ingress *v1beta1.Ingress, backend *v1beta1.IngressBackend, extra ...string) string {
	return strings.Join(append([]string{
		ingress.Namespace,
		ingress.Name,
		backend.ServiceName},
		extra...), ".")
}
