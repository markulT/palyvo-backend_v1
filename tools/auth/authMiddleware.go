package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"palyvoua/internal/models"
	"strings"
)

func AuthMiddleware(userRepo userRepo) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(401, gin.H{"error":"Unauthorized"})
			c.Abort()
			return
		}
		accessToken := authHeader[7:]
		if _, err := Validate(accessToken);err!=nil {
			c.JSON(401, gin.H{"error":"Invalid token"})
			c.Abort()
			return
		}
		userEmail, err := GetSubject(accessToken)
		if err != nil {
			c.JSON(401, gin.H{"error":"Error extracting subject from token (invalid token)"})
			c.Abort()
			return
		}

		user, err := userRepo.GetUserByEmail(userEmail)
		if err != nil {
			c.JSON(404, gin.H{"error":"No such user"})
			c.Abort()
			return
		}

		c.Set("userEmail", userEmail)
		c.Set("user", user)
		c.Next()
	}
}

type userRepo interface {
	GetUserByEmail(email string) (models.User, error)
}

type roleRepo interface {
	GetRoleByID(uuid.UUID) (models.Role, error)
}

func RoleMiddleware(requiredAuthorityLevel int, userRepo userRepo, roleRepo roleRepo) gin.HandlerFunc {
	return func(c *gin.Context) {
		userEmail, exists := c.Get("userEmail")

		if !exists {
			c.Set("authorized", false)
			c.Set("authorityLevel", -1)
			c.JSON(403, gin.H{"error":"Error authorizing user"})
			c.Abort()
			return
		}

		user,err := userRepo.GetUserByEmail(userEmail.(string))
		if err != nil {
			c.Set("authorized", false)
			c.Set("authorityLevel", -1)
			c.JSON(403, gin.H{"error":"Error authorizing user"})
			c.Abort()
			return
		}
		role, err := roleRepo.GetRoleByID(user.Role)
		if role.AuthorityLevel < requiredAuthorityLevel {
			c.Set("authorized", false)
			c.Set("authorityLevel", -1)
			c.JSON(403, gin.H{"error":"You have no authority for this (FORBIDDEN)"})
			c.Abort()
			return
		}
		c.Set("authorized", true)
		c.Set("authorityLevel", role.AuthorityLevel)
		c.Next()

	}
}
