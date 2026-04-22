package middleware

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/thomas-illiet/ai-bridge/database"
	"github.com/thomas-illiet/ai-bridge/models"
)

type whitelistCache struct {
	mu         sync.RWMutex
	nets       []*net.IPNet
	loadedAt   time.Time
	hasEntries bool
}

var cache = &whitelistCache{}

// InvalidateIPCache forces the next request to reload the whitelist from DB.
func InvalidateIPCache() {
	cache.mu.Lock()
	cache.loadedAt = time.Time{}
	cache.mu.Unlock()
}

func (wc *whitelistCache) load() {
	var entries []models.IPAllowlist
	database.DB.Where("enabled = true").Find(&entries)

	nets := make([]*net.IPNet, 0, len(entries))
	for _, e := range entries {
		cidr := e.CIDR
		if !strings.Contains(cidr, "/") {
			cidr += "/32"
		}
		_, ipNet, err := net.ParseCIDR(cidr)
		if err == nil {
			nets = append(nets, ipNet)
		}
	}

	wc.mu.Lock()
	wc.nets = nets
	wc.hasEntries = len(nets) > 0
	wc.loadedAt = time.Now()
	wc.mu.Unlock()
}

func (wc *whitelistCache) isAllowed(ip net.IP) bool {
	wc.mu.RLock()
	stale := time.Since(wc.loadedAt) > 30*time.Second
	wc.mu.RUnlock()

	if stale {
		wc.load()
	}

	wc.mu.RLock()
	defer wc.mu.RUnlock()

	if !wc.hasEntries {
		return true
	}
	for _, n := range wc.nets {
		if n.Contains(ip) {
			return true
		}
	}
	return false
}

// parseCIDRs parses a comma-separated list of CIDRs/IPs into []*net.IPNet.
func parseCIDRs(list string) []*net.IPNet {
	var nets []*net.IPNet
	for _, raw := range strings.Split(list, ",") {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		if !strings.Contains(raw, "/") {
			raw += "/32"
		}
		_, n, err := net.ParseCIDR(raw)
		if err == nil {
			nets = append(nets, n)
		}
	}
	return nets
}

// isTrusted reports whether ip falls within any of the trusted proxy CIDRs.
func isTrusted(ip net.IP, trusted []*net.IPNet) bool {
	for _, n := range trusted {
		if n.Contains(ip) {
			return true
		}
	}
	return false
}

// realClientIP resolves the actual client IP from a request by walking
// X-Forwarded-For right-to-left, skipping IPs that belong to trusted proxies.
// Falls back to X-Real-IP then RemoteAddr.
func realClientIP(r *http.Request, trusted []*net.IPNet) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		// Walk from right to left: skip trusted proxy IPs.
		for i := len(parts) - 1; i >= 0; i-- {
			candidate := strings.TrimSpace(parts[i])
			ip := net.ParseIP(candidate)
			if ip == nil {
				continue
			}
			if !isTrusted(ip, trusted) {
				return candidate
			}
		}
	}

	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// IPWhitelist returns a Gin middleware that enforces the IP whitelist stored in DB.
// trustedProxies is a comma-separated list of CIDRs covering reverse proxies
// (e.g. "10.0.0.0/8,172.16.0.0/12"). If the whitelist table is empty, all IPs are allowed.
func IPWhitelist(trustedProxies string) gin.HandlerFunc {
	trusted := parseCIDRs(trustedProxies)

	return func(c *gin.Context) {
		ipStr := realClientIP(c.Request, trusted)
		parsed := net.ParseIP(ipStr)
		if parsed == nil || !cache.isAllowed(parsed) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "IP not whitelisted"})
			return
		}
		c.Next()
	}
}
