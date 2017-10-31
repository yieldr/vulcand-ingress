package vulcan

import (
	"bytes"
	"fmt"
)

func CreateRoute(host, path string) string {
	var b bytes.Buffer
	var and bool
	if host != "" {
		fmt.Fprintf(&b, "Host(`%s`)", host)
		and = true
	}
	if path != "" {
		if and {
			fmt.Fprint(&b, " && ")
		}
		fmt.Fprintf(&b, "PathRegexp(`%s`)", path)
	}
	return b.String()
}
