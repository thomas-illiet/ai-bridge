package ctxutil

import "context"

type clientIPKey struct{}

// WithClientIP returns a new context carrying the resolved client IP.
func WithClientIP(ctx context.Context, ip string) context.Context {
	return context.WithValue(ctx, clientIPKey{}, ip)
}

// ClientIPFromContext returns the client IP stored by WithClientIP, or "".
func ClientIPFromContext(ctx context.Context) string {
	ip, _ := ctx.Value(clientIPKey{}).(string)
	return ip
}
