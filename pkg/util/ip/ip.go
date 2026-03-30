// Package ip provides utilities for obtaining local and remote IP addresses.
package ip

import (
	"net"
	"net/http"
)

const (
	XForwardedFor = "X-Forwarded-For"
	XRealIP       = "X-Real-IP"
	XClientIP     = "x-client-ip"
)

// GetLocalIP returns the first non-loopback IPv4 address of the host.
// Falls back to "127.0.0.1" on error.
func GetLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "127.0.0.1"
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return "127.0.0.1"
}

// RemoteIP extracts the client IP from an HTTP request, checking common
// proxy headers (X-Client-IP, X-Real-IP, X-Forwarded-For) before falling
// back to RemoteAddr.
func RemoteIP(req *http.Request) string {
	remoteAddr := req.RemoteAddr
	if ip := req.Header.Get(XClientIP); ip != "" {
		remoteAddr = ip
	} else if ip := req.Header.Get(XRealIP); ip != "" {
		remoteAddr = ip
	} else if ip = req.Header.Get(XForwardedFor); ip != "" {
		remoteAddr = ip
	} else {
		remoteAddr, _, _ = net.SplitHostPort(remoteAddr)
	}
	if remoteAddr == "::1" {
		remoteAddr = "127.0.0.1"
	}
	return remoteAddr
}
