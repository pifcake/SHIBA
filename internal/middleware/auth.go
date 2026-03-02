package middleware

import (
	"strings"

	"SHIBA/internal/domain"
	jwtpkg "SHIBA/pkg/jwt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"SHIBA/pkg/response"
)

const (
	UserIDKey   = "user_id"
	UserRoleKey = "user_role"
)

func Auth(jwtManager *jwtpkg.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "missing authorization header")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			response.Unauthorized(c, "invalid authorization header format")
			c.Abort()
			return
		}

		claims, err := jwtManager.ParseAccessToken(parts[1])
		if err != nil {
			response.Unauthorized(c, "invalid or expired token")
			c.Abort()
			return
		}

		c.Set(UserIDKey, claims.UserID)
		c.Set(UserRoleKey, claims.Role)
		c.Next()
	}
}

func RequireRoles(roles ...domain.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleStr, exists := c.Get(UserRoleKey)
		if !exists {
			response.Unauthorized(c, "not authenticated")
			c.Abort()
			return
		}

		role := domain.UserRole(roleStr.(string))
		for _, r := range roles {
			if role == r {
				c.Next()
				return
			}
		}

		response.Forbidden(c)
		c.Abort()
	}
}

func GetUserID(c *gin.Context) uuid.UUID {
	id, _ := c.Get(UserIDKey)
	if uid, ok := id.(uuid.UUID); ok {
		return uid
	}
	return uuid.Nil
}

func GetUserRole(c *gin.Context) domain.UserRole {
	role, _ := c.Get(UserRoleKey)
	if r, ok := role.(string); ok {
		return domain.UserRole(r)
	}
	return ""
}
