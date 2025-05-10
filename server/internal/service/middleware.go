package service

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/liriquew/secret_storage/server/internal/lib/jwt"
)

func (s *Service) ShamirRequired(c *gin.Context) {
	if s.repository == nil {
		c.AbortWithStatusJSON(http.StatusTeapot, gin.H{
			"type": "storage is encrypted now",
		})
		return
	}

	c.Next()
}

func (s *Service) AuthRequired(c *gin.Context) {
	headerVal := c.GetHeader("Authorization")

	token, found := strings.CutPrefix(headerVal, "Bearer ")
	if !found {
		s.log.Error("bad Bearer scheme", slog.String("token", token))
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"type": "bad header value (Bearer scheme required)"})
		return
	}

	username, err := jwt.Validate(token)
	if err != nil {
		s.log.Error("bad token signature", slog.String("token", token))
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"type": "bad jwt"})
		return
	}

	c.Set(usernameKey, username)

	c.Next()
}
