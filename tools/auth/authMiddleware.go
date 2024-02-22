package auth

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"palyvoua/internal/models"
	"palyvoua/internal/repository"
	"strings"
)

type AuthBody struct {
	user *models.User
	role *models.Role
}

func (ab *AuthBody) setUser(u *models.User) error {
	ab.user = u
	return nil
}

func (ab *AuthBody) setRole(r *models.Role) error {
	ab.role = r
	return nil
}

func (ab *AuthBody) GetUser() *models.User {
	return ab.user
}

func (ab *AuthBody) GetRole() *models.Role {
	return ab.role
}

func AuthMiddleware(userRepo repository.UserRepo, roleRepo roleRepo) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(401, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}
		accessToken := authHeader[7:]
		if _, err := Validate(accessToken); err != nil {
			fmt.Println("auth 2")
			c.JSON(401, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}
		userEmail, err := GetSubject(accessToken)
		if err != nil {

			c.JSON(401, gin.H{"error": "Error extracting subject from token (invalid token)"})
			c.Abort()
			return
		}

		user, err := userRepo.GetUserByEmail(context.Background(), userEmail)
		if err != nil {

			c.JSON(404, gin.H{"error": "No such user"})
			c.Abort()
			return
		}
		fmt.Println(user.Role)
		role, err := roleRepo.GetRoleByID(user.Role)
		if err != nil {

			c.JSON(403, gin.H{"error": "You have no authority for this (FORBIDDEN)"})
			c.Abort()
			return
		}

		authBody := AuthBody{}
		err = authBody.setUser(&user)
		if err != nil {

			c.JSON(400, gin.H{"error": "Error authorising user"})
			c.Abort()
			return
		}
		err = authBody.setRole(&role)
		if err != nil {
			c.JSON(400, gin.H{"error": "Error authorising user"})
			c.Abort()
			return
		}
		c.Set("authBody", authBody)

		c.Next()
	}
}

type userRepo interface {
	GetUserByEmail(ctx context.Context, email string) (models.User, error)
}

type roleRepo interface {
	GetRoleByID(uuid.UUID) (models.Role, error)
}

func RoleMiddleware(requiredAuthorityLevel int, userRepo repository.UserRepo, roleRepo roleRepo) gin.HandlerFunc {
	return func(c *gin.Context) {
		authBodyField, exists := c.Get("authBody")

		if !exists {
			c.Set("authorized", false)
			c.Set("authorityLevel", -1)
			c.JSON(403, gin.H{"error": "Error authorizing user: " + "does not exist"})
			c.Abort()
			return
		}

		authBody, ok := authBodyField.(AuthBody)

		if !ok {
			c.Set("authorized", false)
			c.Set("authorityLevel", -1)
			c.JSON(403, gin.H{"error": "Error authorizing user: " + "failed to cast"})
			c.Abort()
			return
		}

		user, err := userRepo.GetUserByEmail(context.Background(), authBody.user.Email)
		if err != nil {
			c.Set("authorized", false)
			c.Set("authorityLevel", -1)
			c.JSON(403, gin.H{"error": "Error authorizing user: " + err.Error()})
			c.Abort()
			return
		}
		role, err := roleRepo.GetRoleByID(user.Role)
		if role.AuthorityLevel < requiredAuthorityLevel {
			c.Set("authorized", false)
			c.Set("authorityLevel", -1)
			c.JSON(403, gin.H{"error": "You have no authority for this (FORBIDDEN)"})
			c.Abort()
			return
		}
		c.Set("authorized", true)
		c.Set("authorityLevel", role.AuthorityLevel)
		c.Next()

	}
}
