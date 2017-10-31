package vulcan

import (
	"testing"

	"k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestCreateURL(t *testing.T) {
	for expected, backend := range map[string]*v1beta1.IngressBackend{
		"http://foo:80": &v1beta1.IngressBackend{
			ServiceName: "foo",
			ServicePort: intstr.FromInt(80),
		},
		"http://bar:443": &v1beta1.IngressBackend{
			ServiceName: "bar",
			ServicePort: intstr.FromString("443"),
		},
	} {
		url := CreateURL(backend)
		if url != expected {
			t.Errorf("Unexpected URL %q from service %q and port %q", url, backend.ServiceName, backend.ServicePort)
		}
	}
}
