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
	mu        sync.RWMutex
	nets      []*net.IPNet
	loadedAt  time.Time
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
	var entries []models.IPWhitelistEntry
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

// IPWhitelist returns a Gin middleware that enforces the IP whitelist stored in DB.
// If the whitelist table is empty, all IPs are allowed.
func IPWhitelist() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := clientIP(c.Request)
		parsed := net.ParseIP(ip)
		if parsed == nil || !cache.isAllowed(parsed) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "IP not whitelisted"})
			return
		}
		c.Next()
	}
}

func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if parts := strings.Split(xff, ","); len(parts) > 0 {
			return strings.TrimSpace(parts[0])
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
