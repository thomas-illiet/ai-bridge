package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/coder/aibridge"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/thomas-illiet/ai-bridge/config"
	"github.com/thomas-illiet/ai-bridge/models"
	"github.com/thomas-illiet/ai-bridge/services"
)

const ctxUserKey = "user"

// Role constants — re-exported from models for convenience.
const (
	RoleAdmin   = models.RoleAdmin
	RoleManager = models.RoleManager
	RoleUser    = models.RoleUser
	RoleService = models.RoleService
	RoleNone    = models.RoleNone
)

type OIDCClaims struct {
	jwt.RegisteredClaims
	PreferredUsername string `json:"preferred_username"`
	Email             string `json:"email"`
	GivenName         string `json:"given_name"`
	FamilyName        string `json:"family_name"`
}

func resolveJWKSUrl(ctx context.Context, cfg *config.Config) (string, error) {
	if u := cfg.JWKSUrl(); u != "" {
		return u, nil
	}
	discoveryURL := strings.TrimRight(cfg.OIDCIssuerURL, "/") + "/.well-known/openid-configuration"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, discoveryURL, nil)
	if err != nil {
		return "", err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("OIDC discovery failed: %w", err)
	}
	defer resp.Body.Close()
	var doc struct {
		JWKSURI string `json:"jwks_uri"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&doc); err != nil {
		return "", fmt.Errorf("OIDC discovery decode failed: %w", err)
	}
	if doc.JWKSURI == "" {
		return "", fmt.Errorf("OIDC discovery returned empty jwks_uri")
	}
	return doc.JWKSURI, nil
}

func JWTAuth(cfg *config.Config) gin.HandlerFunc {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	jwksURL, err := resolveJWKSUrl(ctx, cfg)
	if err != nil {
		panic("failed to resolve JWKS URL from OIDC provider: " + err.Error())
	}

	jwks, err := keyfunc.NewDefaultCtx(ctx, []string{jwksURL})
	if err != nil {
		panic("failed to fetch JWKS from OIDC provider: " + err.Error())
	}

	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid Authorization header"})
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

		// Try OIDC JWT first.
		oidcClaims := &OIDCClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, oidcClaims, jwks.KeyfuncCtx(c.Request.Context()),
			jwt.WithIssuer(cfg.OIDCIssuerURL),
			jwt.WithExpirationRequired(),
		)
		if err == nil && token.Valid {
			registered, err := services.GetOrCreateUser(
				oidcClaims.Subject,
				oidcClaims.PreferredUsername,
				oidcClaims.Email,
			)
			if err != nil || registered == nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "user registration failed"})
				return
			}
			c.Set(ctxUserKey, &models.AuthUser{
				ID:                registered.ID,
				Username:          registered.Username,
				Email:             registered.Email,
				FirstName:         oidcClaims.GivenName,
				LastName:          oidcClaims.FamilyName,
				Roles:             []string{registered.Role},
				PreferredUsername: registered.Username,
			})
			c.Next()
			return
		}

		// Fall back to PAT validation.
		patClaims := &services.PATClaims{}
		jwt.NewParser().ParseUnverified(tokenStr, patClaims) //nolint:errcheck
		if jti := patClaims.ID; jti != "" {
			record, err := services.LookupAndVerify(jti, tokenStr, cfg.TokenSecret)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
				return
			}
			if record != nil {
				// Look up DB role — PAT users are not automatically granted any role.
				registered, err := services.GetUserByID(record.UserID)
				role := RoleNone
				if err == nil && registered != nil {
					role = registered.Role
				}
				c.Set(ctxUserKey, &models.AuthUser{
					ID:                record.UserID,
					PreferredUsername: record.UserID,
					Roles:             []string{role},
				})
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
	}
}

func GetUser(c *gin.Context) *models.AuthUser {
	u, _ := c.Get(ctxUserKey)
	user, _ := u.(*models.AuthUser)
	return user
}

func RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := GetUser(c)
		if user == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthenticated"})
			return
		}
		for _, r := range user.Roles {
			if r == role {
				c.Next()
				return
			}
		}
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
	}
}

// RequireAnyRole allows access if the user has at least one of the given roles.
func RequireAnyRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := GetUser(c)
		if user == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthenticated"})
			return
		}
		for _, required := range roles {
			for _, r := range user.Roles {
				if r == required {
					c.Next()
					return
				}
			}
		}
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
	}
}

// AIBridgeActor injects the authenticated user ID into the request context
// as an aibridge actor, enabling per-user usage recording.
func AIBridgeActor() gin.HandlerFunc {
	return func(c *gin.Context) {
		user := GetUser(c)
		if user != nil {
			ctx := aibridge.AsActor(c.Request.Context(), user.ID, nil)
			c.Request = c.Request.WithContext(ctx)
		}
		c.Next()
	}
}
