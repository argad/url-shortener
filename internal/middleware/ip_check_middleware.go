package middleware

import (
	"net"
	"net/http"
)

// CheckTrustedSubnet is a middleware that checks if the request comes from a trusted subnet.
func CheckTrustedSubnet(trustedSubnet string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// If trusted subnet is not configured, deny all requests.
			if trustedSubnet == "" {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			_, ipNet, err := net.ParseCIDR(trustedSubnet)
			if err != nil {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			realIP := r.Header.Get("X-Real-IP")
			if realIP == "" {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			ip := net.ParseIP(realIP)
			if ip == nil {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			if !ipNet.Contains(ip) {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
