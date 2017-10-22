package version

import (
	"fmt"
	"runtime"
)

var Version = "v0.0.0"

func LongVersion() string {
	return fmt.Sprintf(
		"Vulcand Ingress Controller %s (%s_%s)",
		Version,
		runtime.GOOS,
		runtime.GOARCH)
}
