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

type KeycloakClaims struct {
	jwt.RegisteredClaims
	PreferredUsername string         `json:"preferred_username"`
	Email             string         `json:"email"`
	GivenName         string         `json:"given_name"`
	FamilyName        string         `json:"family_name"`
	RealmAccess       realmAccess    `json:"realm_access"`
}

type realmAccess struct {
	Roles []string `json:"roles"`
}

func JWTAuth(cfg *config.Config) gin.HandlerFunc {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	jwks, err := keyfunc.NewDefaultCtx(ctx, []string{cfg.JWKSUrl()})
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
			c.Set(ctxUserKey, &models.User{
				ID:                keycloakClaims.Subject,
				Username:          keycloakClaims.PreferredUsername,
				Email:             keycloakClaims.Email,
				FirstName:         keycloakClaims.GivenName,
				LastName:          keycloakClaims.FamilyName,
				Roles:             keycloakClaims.RealmAccess.Roles,
				PreferredUsername: keycloakClaims.PreferredUsername,
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
				c.Set(ctxUserKey, &models.User{ID: record.UserID, PreferredUsername: record.UserID})
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
