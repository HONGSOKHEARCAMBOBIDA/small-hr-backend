package middleware

import (
	"mysql/config"
	"mysql/constant/share"
	"mysql/helper"
	"mysql/model"
	"mysql/utils"
	"net/http" // HTTP status codes
	"strings"

	"github.com/gin-gonic/gin" // Gin web framework
	"github.com/golang-jwt/jwt/v5"
)

// AuthMiddleware creates a Gin middleware function for JWT authentication
func AuthMiddleware() gin.HandlerFunc {
	// Return the actual middleware handler function
	return func(c *gin.Context) {

		// 1. Check for Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// If no Authorization header, abort with 401 Unauthorized
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Authorization header missing"})
			c.Abort() // Stop processing the request
			return
		}

		// 2. Validate Authorization header format
		// Expected format: "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			// If format is invalid, abort with 401
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid authorization format"})
			c.Abort()
			return
		}

		// Extract the token part (after "Bearer ")
		tokenString := parts[1]

		// 3. Parse and validate the JWT
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Verify the token's signing algorithm is HMAC
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			// Return the secret key for verification
			return utils.Jwtkey, nil
		})

		// Check if token is invalid or parsing failed
		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid or expired token"})
			c.Abort()
			return
		}

		// 4. If token is valid, extract claims and set them in context
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			// Set user_id from claims to Gin context
			c.Set("user_id", claims["user_id"])
			// Set role_id from claims to Gin context
			c.Set("role_id", claims["role_id"])
			// You can add more claims here as needed
		}

		userID, ok := helper.GetUserID(c)

		if !ok {

			share.ResponseError(c, http.StatusUnauthorized, "Please Login")
			c.Abort()
			return
		}

		var user model.User

		if err := config.DB.Where("id = ? AND is_active = ?", userID, 1).First(&user).Error; err != nil {

			c.AbortWithStatusJSON(403, gin.H{"message": "User is not active"})

			return
		}

		c.Next()
	}
}
