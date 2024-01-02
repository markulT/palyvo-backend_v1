package auth

import (
	"github.com/gin-gonic/gin"
	"golearn/repository"
	"strings"
)

func SubscriptionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(401, gin.H{"error":"Unauthorized"})
			c.Abort()
			return
		}
		accessToken := authHeader[7:]

		userEmail, err := GetSubject(accessToken)
		if err != nil {
			c.JSON(401, gin.H{"error":"Unauthorized"})
			c.Abort()
			return
		}

		userFromDB, err := repository.GetUserByEmail(userEmail)
		if err != nil {
			c.JSON(401, gin.H{"error":"Unauthorized"})
			c.Abort()
			return
		}

		if idValid:=verifySubscriptionID(userFromDB.SubscriptionID);!idValid {
			c.JSON(403, gin.H{"error":"Forbidden", "message":"The user is not subscribed"})
			c.Abort()
			return
		}

		c.Next()
	}
}
// verifySubscriptionID - returns true if subID is valid, and false if it's not
func verifySubscriptionID(subID string) bool {
	return subID != ""
}

func CheckSubLevel(email string, requiredSubLevel int) (bool, error) {
	subLevel, err := repository.GetUserSubLevelbyEmail(email)

	if err != nil {
		return false, err
	}
	if subLevel < requiredSubLevel {
		return false, nil
	}
	return true, nil
}

func SubLevelMiddleware(requiredSubLevel int) gin.HandlerFunc {
	return func(c *gin.Context) {
		userEmail, exists := c.Get("userEmail")
		if !exists {
			c.JSON(500, gin.H{"error":"Internal server error"})
			c.Abort()
			return
		}
		allowed, err := CheckSubLevel(userEmail.(string), requiredSubLevel)
		if err != nil {

			c.JSON(500, gin.H{"error":err.Error()})
			c.Abort()
			return
		}
		if !allowed {
			c.JSON(403, gin.H{"error":"Users sub level is not enough"})
			c.Abort()
			return
		}
		c.Next()
	}
}
