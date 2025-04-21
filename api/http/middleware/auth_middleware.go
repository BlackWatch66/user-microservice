package middleware

import (
	"log"
	"net/http"
	"strings"

	"github.com/blackwatch66/user-microservice/config"
	"github.com/blackwatch66/user-microservice/internal/auth"
	"github.com/gin-gonic/gin"
)

const (
	AuthorizationHeaderKey = "Authorization"
	BearerSchema           = "Bearer"
    UserContextKey         = "userClaims" // Context key to store user claims
)

// AuthMiddleware 创建一个 Gin 中间件用于 JWT 认证
func AuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader(AuthorizationHeaderKey)
		log.Printf("Received Authorization header: %s", authHeader)
		
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			return
		}

		// 处理可能没有正确分隔的情况
		var tokenString string
		if strings.HasPrefix(strings.ToLower(authHeader), strings.ToLower(BearerSchema)) {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 {
				tokenString = parts[1]
			} else {
				// Bearer后面直接跟着Token，没有空格分隔
				tokenString = strings.TrimPrefix(authHeader, BearerSchema)
			}
		} else {
			// 没有Bearer前缀，假设整个值就是token
			tokenString = authHeader
		}
		
		log.Printf("Extracted token: %s", tokenString)

		claims, err := auth.ValidateJWT(tokenString, cfg.JWTSecret)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token", "details": err.Error()})
			return
		}

        // 可选：检查 Redis 中的 token 标识符
        // ... (类似 service 中的 ValidateToken 逻辑)

		// 将用户信息存入 Gin 的 Context 中，方便后续 Handler 使用
		c.Set(UserContextKey, claims)
		log.Printf("Authentication successful for user ID: %d", claims.UserID)

		c.Next()
	}
}

// GetUserClaims 从 Gin Context 中获取用户信息
func GetUserClaims(c *gin.Context) (*auth.Claims, bool) {
    claims, exists := c.Get(UserContextKey)
    if !exists {
        return nil, false
    }
    userClaims, ok := claims.(*auth.Claims)
    return userClaims, ok
} 