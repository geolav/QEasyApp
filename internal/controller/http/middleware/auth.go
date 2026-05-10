package middleware

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func Auth(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// сначала пробуем заголовок
		authHeader := c.GetHeader("Authorization")

		var tokenStr string

		if authHeader != "" {
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token format"})
				c.Abort()
				return
			}
			tokenStr = parts[1]
		} else {
			// если заголовка нет — пробуем query параметр ?token=...
			tokenStr = c.Query("token")
			if tokenStr == "" {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
				c.Abort()
				return
			}
		}

		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		claims := token.Claims.(jwt.MapClaims)
		userID, ok := claims["user_id"].(string)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
			c.Abort()
			return
		}
		log.Printf("middleware userID: '%s'", userID)
		c.Set("user_id", userID)
		c.Next()
	}
}
