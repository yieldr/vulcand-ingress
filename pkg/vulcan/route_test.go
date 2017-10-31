package vulcan

import "testing"

func TestCreateRoute(t *testing.T) {
	for expected, test := range map[string]struct {
		host string
		path string
	}{
		"PathRegexp(`/hello`)":                        {"", "/hello"},
		"Host(`example.com`)":                         {"example.com", ""},
		"Host(`example.com`) && PathRegexp(`/hello`)": {"example.com", "/hello"},
	} {
		route := CreateRoute(test.host, test.path)
		if route != expected {
			t.Errorf("Unexpected route %q from host: %q and path %q", route, test.host, test.path)
		}
	}
}
