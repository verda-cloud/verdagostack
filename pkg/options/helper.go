package options

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

const (
	_   = iota
	KiB = 1 << (10 * iota)
	MiB
	GiB
	TiB
)

// Join concatenates prefixes with "." separators.
// If the result is non-empty, a trailing "." is appended.
func Join(prefixes ...string) string {
	joined := strings.Join(prefixes, ".")
	if joined != "" {
		joined += "."
	}
	return joined
}

// ValidateAddress validates that addr is in a valid :port or host:port format.
func ValidateAddress(addr string) error {
	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return fmt.Errorf("%q is not in a valid format (:port or ip:port): %w", addr, err)
	}
	if host != "" && net.ParseIP(host) == nil {
		return fmt.Errorf("%q is not a valid IP address", host)
	}
	port, err := strconv.Atoi(portStr)
	if err != nil || port < 0 || port > 65535 {
		return fmt.Errorf("%q is not a valid port number", portStr)
	}
	return nil
}

// GetLocalIP returns the first non-loopback IPv4 address of the host.
func GetLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "127.0.0.1"
	}
	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				return ipNet.IP.String()
			}
		}
	}
	return "127.0.0.1"
}
