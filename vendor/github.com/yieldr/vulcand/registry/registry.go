package registry

import (
	"github.com/vulcand/vulcand/plugin"
	"github.com/vulcand/vulcand/plugin/cbreaker"
	"github.com/vulcand/vulcand/plugin/connlimit"
	"github.com/vulcand/vulcand/plugin/ratelimit"
	"github.com/vulcand/vulcand/plugin/rewrite"
	"github.com/vulcand/vulcand/plugin/trace"
	"github.com/yieldr/vulcand/plugin/auth"
	"github.com/yieldr/vulcand/plugin/oauth2"
)

func GetRegistry() (*plugin.Registry, error) {
	r := plugin.NewRegistry()

	specs := []*plugin.MiddlewareSpec{
		cbreaker.GetSpec(),
		connlimit.GetSpec(),
		ratelimit.GetSpec(),
		rewrite.GetSpec(),
		trace.GetSpec(),
		oauth2.GetSpec(),
		auth.GetSpec(),
		auth.GetLegacySpec(),
	}

	for _, spec := range specs {
		if err := r.AddSpec(spec); err != nil {
			return nil, err
		}
	}
	return r, nil
}
