package annotations

import (
	"testing"

	"k8s.io/api/extensions/v1beta1"
)

func TestGetString(t *testing.T) {

	annotations := map[string]string{
		TrustForwardHeader:  "true",
		ReadTimeout:         "5s",
		DialTimeout:         "7s",
		TLSHandshakeTimeout: "10s",
		KeepAlive:           "30s",
		MaxIdleConnsPerHost: "12",
	}

	ingress := &v1beta1.Ingress{}
	ingress.SetAnnotations(annotations)

	for annotation, expected := range annotations {
		actual := GetString(ingress, annotation)
		if actual != expected {
			t.Errorf("Unexpected annotation value %q, expected %q", actual, expected)
		}
	}
}

func TestGetMiddleware(t *testing.T) {

	annotations := map[string]string{
		"ingress.kubernetes.io/middleware.ratelimit": `{"PeriodSeconds":1,"Burst":3,"Variable":"client.ip","Requests":1}`,
		"ingress.kubernetes.io/middleware.connlimit": `{"Connections":3,"Variable":"client.ip"}`,
	}

	ingress := &v1beta1.Ingress{}
	ingress.SetAnnotations(annotations)

	middleware := GetMiddleware(ingress)

	for _, name := range []string{
		"ratelimit",
		"connlimit",
	} {
		t.Run(name, func(t *testing.T) {
			if middleware[name] == "" {
				t.Error("Unexpected empty middleware")
			}
		})
	}
}
