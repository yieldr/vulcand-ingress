package annotations

import (
	"regexp"
	"strconv"

	"k8s.io/api/extensions/v1beta1"
)

const (
	// Backend related annotations
	ReadTimeout         = "ingress.kubernetes.io/read-timeout"
	DialTimeout         = "ingress.kubernetes.io/dial-timeout"
	TLSHandshakeTimeout = "ingress.kubernetes.io/tls-handshake-timeout"
	KeepAlive           = "ingress.kubernetes.io/keepalive"
	MaxIdleConnsPerHost = "ingress.kubernetes.io/max-idle-connections-per-host"

	// Frontend related annotations
	TrustForwardHeader = "ingress.kubernetes.io/trust-forward-header"
	PassHostHeader     = "ingress.kubernetes.io/pass-host-header"
	MaxBodyBytes       = "ingress.kubernetes.io/max-body-bytes"
	MaxMemBodyBytes    = "ingress.kubernetes.io/max-mem-body-bytes"
	FailoverPredicate  = "ingress.kubernetes.io/failover-predicate"
	Hostname           = "ingress.kubernetes.io/hostname"
)

var middlewareRegexp = regexp.MustCompile(`ingress.kubernetes.io/middleware\.(.*)`)

func String(a string) string {
	return string(a)
}

func GetString(obj *v1beta1.Ingress, a string) string {
	return String(obj.Annotations[a])
}

func Int(a string) int {
	i, err := strconv.Atoi(string(a))
	if err != nil {
		return 0
	}
	return i
}

func GetInt(obj *v1beta1.Ingress, a string) int {
	return Int(obj.Annotations[a])
}

func Bool(a string) bool {
	b, err := strconv.ParseBool(string(a))
	if err != nil {
		return false
	}
	return b
}

func GetBool(obj *v1beta1.Ingress, a string) bool {
	return Bool(obj.Annotations[a])
}

func GetMiddleware(obj *v1beta1.Ingress) map[string]string {
	middleware := make(map[string]string)
	for key, value := range obj.Annotations {
		match := middlewareRegexp.FindStringSubmatch(key)
		if len(match) == 2 {
			middleware[match[1]] = value
		}
	}
	return middleware
}
