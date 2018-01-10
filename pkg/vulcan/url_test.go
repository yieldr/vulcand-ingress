package vulcan

import (
	"testing"

	"k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestCreateURL(t *testing.T) {
	ingress := &v1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ingress",
			Namespace: "namespace",
		},
	}
	for expected, backend := range map[string]*v1beta1.IngressBackend{
		"http://foo.namespace:80": &v1beta1.IngressBackend{
			ServiceName: "foo",
			ServicePort: intstr.FromInt(80),
		},
		"http://bar.namespace:443": &v1beta1.IngressBackend{
			ServiceName: "bar",
			ServicePort: intstr.FromString("443"),
		},
	} {
		url := CreateURL(ingress, backend)
		if url != expected {
			t.Errorf("Unexpected URL %q from service %q and port %q", url, backend.ServiceName, backend.ServicePort)
		}
	}
}
