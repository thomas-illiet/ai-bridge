package middleware

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/thomas-illiet/ai-bridge/internal/ctxutil"
	"github.com/thomas-illiet/ai-bridge/internal/database"
	"github.com/thomas-illiet/ai-bridge/internal/models"
)

type firewallEntry struct {
	net    *net.IPNet
	action string
}

type firewallCache struct {
	mu       sync.RWMutex
	rules    []firewallEntry
	loadedAt time.Time
}

var fwCache = &firewallCache{}

// InvalidateFirewallCache forces the next request to reload the firewall rules from DB.
func InvalidateFirewallCache() {
	fwCache.mu.Lock()
	fwCache.loadedAt = time.Time{}
	fwCache.mu.Unlock()
}

func (fc *firewallCache) load() {
	var rules []models.FirewallRule
	database.DB.Where("enabled = true").Order("priority asc").Find(&rules)

	entries := make([]firewallEntry, 0, len(rules))
	for _, r := range rules {
		cidr := r.CIDR
		if !strings.Contains(cidr, "/") {
			cidr += "/32"
		}
		_, ipNet, err := net.ParseCIDR(cidr)
		if err == nil {
			entries = append(entries, firewallEntry{net: ipNet, action: r.Action})
		}
	}

	fc.mu.Lock()
	fc.rules = entries
	fc.loadedAt = time.Now()
	fc.mu.Unlock()
}

func (fc *firewallCache) evaluate(ip net.IP) bool {
	fc.mu.RLock()
	stale := time.Since(fc.loadedAt) > 10*time.Minute
	fc.mu.RUnlock()

	if stale {
		fc.load()
	}

	fc.mu.RLock()
	defer fc.mu.RUnlock()

	for _, r := range fc.rules {
		if r.net.Contains(ip) {
			return r.action == "allow"
		}
	}
	return true
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

// InjectClientIP is a Gin middleware that resolves the real client IP (honoring
// trusted proxies) and stores it in the request context for downstream use.
func InjectClientIP(trustedProxies string) gin.HandlerFunc {
	trusted := parseCIDRs(trustedProxies)

	return func(c *gin.Context) {
		ip := realClientIP(c.Request, trusted)
		ctx := ctxutil.WithClientIP(c.Request.Context(), ip)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

// Firewall returns a Gin middleware that enforces firewall rules stored in DB.
// Rules are evaluated in priority order (ASC); first matching CIDR wins.
// If no rule matches, the request is allowed (open by default).
// trustedProxies is a comma-separated list of CIDRs covering reverse proxies.
func Firewall(trustedProxies string) gin.HandlerFunc {
	trusted := parseCIDRs(trustedProxies)

	return func(c *gin.Context) {
		ipStr := realClientIP(c.Request, trusted)
		parsed := net.ParseIP(ipStr)
		if parsed == nil || !fwCache.evaluate(parsed) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "IP blocked by firewall"})
			return
		}
		c.Next()
	}
}
