package middleware

import (
	"context"
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
	RoleNone    = models.RoleNone
)

type KeycloakClaims struct {
	jwt.RegisteredClaims
	PreferredUsername string `json:"preferred_username"`
	Email             string `json:"email"`
	GivenName         string `json:"given_name"`
	FamilyName        string `json:"family_name"`
}

func JWTAuth(cfg *config.Config) gin.HandlerFunc {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	jwks, err := keyfunc.NewDefaultCtx(ctx, []string{cfg.JWTSUrl()})
	if err != nil {
		panic("failed to fetch JWKS from Keycloak: " + err.Error())
	}

	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid Authorization header"})
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

		// Try Keycloak JWT first.
		keycloakClaims := &KeycloakClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, keycloakClaims, jwks.KeyfuncCtx(c.Request.Context()),
			jwt.WithIssuer(cfg.IssuerURL()),
			jwt.WithExpirationRequired(),
		)
		if err == nil && token.Valid {
			registered, err := services.GetOrCreateUser(
				keycloakClaims.Subject,
				keycloakClaims.PreferredUsername,
				keycloakClaims.Email,
			)
			if err != nil || registered == nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "user registration failed"})
				return
			}
			c.Set(ctxUserKey, &models.User{
				ID:                registered.ID,
				Username:          registered.Username,
				Email:             registered.Email,
				FirstName:         keycloakClaims.GivenName,
				LastName:          keycloakClaims.FamilyName,
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
				c.Set(ctxUserKey, &models.User{
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

func GetUser(c *gin.Context) *models.User {
	u, _ := c.Get(ctxUserKey)
	user, _ := u.(*models.User)
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
