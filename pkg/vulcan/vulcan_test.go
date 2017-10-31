package vulcan

import (
	"testing"

	"k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCreateID(t *testing.T) {
	for expected, test := range map[string]struct {
		ingress *v1beta1.Ingress
		backend *v1beta1.IngressBackend
		extra   []string
	}{
		"namespace.ingress.backend": {
			&v1beta1.Ingress{
				ObjectMeta: v1.ObjectMeta{
					Name:      "ingress",
					Namespace: "namespace",
				},
			},
			&v1beta1.IngressBackend{
				ServiceName: "backend",
			},
			[]string{},
		},
	} {
		id := CreateID(test.ingress, test.backend, test.extra...)
		if id != expected {
			t.Errorf("Unexpected route %q from ingress %q backend %q and extra %q",
				id,
				test.ingress,
				test.backend,
				test.extra)
		}
	}
}
