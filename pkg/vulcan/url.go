package vulcan

import (
	"fmt"

	"k8s.io/api/extensions/v1beta1"
)

func CreateURL(backend *v1beta1.IngressBackend) string {
	return fmt.Sprintf("http://%s:%s",
		backend.ServiceName,
		backend.ServicePort.String())
}
