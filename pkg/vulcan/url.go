package vulcan

import (
	"fmt"

	"k8s.io/api/extensions/v1beta1"
)

func CreateURL(ingress *v1beta1.Ingress, backend *v1beta1.IngressBackend) string {
	return fmt.Sprintf("http://%s.%s:%s",
		backend.ServiceName,
		ingress.Namespace,
		backend.ServicePort.String())
}
