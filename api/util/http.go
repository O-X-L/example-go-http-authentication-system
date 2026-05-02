package util

import (
	"example_api/config"
	"net"
	"net/http"
	"strings"
)

func GetClientIP(r *http.Request) string {
	if !config.IsDeploymentProduction() {
		return "127.0.0.1"
	}

	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		ips := strings.Split(forwarded, ",")
		clientIP := strings.TrimSpace(ips[0])
		if clientIP != "" {
			return clientIP
		}
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}
