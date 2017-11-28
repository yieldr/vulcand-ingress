package vulcan

import (
	"fmt"
	"strings"
)

func CreateRoute(host, path string) string {

	exp := make([]string, 0, 2)

	if host != "" {
		exp = append(exp, fmt.Sprintf("Host(`%s`)", host))
	}

	if path != "" {
		exp = append(exp, fmt.Sprintf("PathRegexp(`%s`)", path))
	}

	return strings.Join(exp, " && ")
}
