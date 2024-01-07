package auth

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"palyvoua/internal/models"
	repository "palyvoua/internal/repository"
)

func UserExists(email string) bool {
	//var count int64
	//utils.DB.Model(&models.User{}).Where("email = ?", email).Count(&count)
	userRepo := repository.NewUserRepo()
	user, _ := userRepo.GetUserByEmail(email)
	fmt.Println(user)
	isEmpty := user == models.User{}
	if isEmpty {
		return false
	}
	return true
}
func GenerateTokens(body map[string]interface{}, c *gin.Context) (Tokens) {
	var tokens Tokens
	accessToken, accessErr := CreateAccessToken(body)
	if accessErr != nil {
		c.JSON(400, gin.H{"message":"Error generating access token"})
	}
	refreshToken, refreshErr := CreateRefreshToken(body)
	if refreshErr != nil {
		c.JSON(400, gin.H{"message":"Error generating refresh token"})
	}
	tokens.AccessToken = accessToken
	tokens.RefreshToken = refreshToken
	return tokens
}
func ComparePasswords(plainPassword string, hashedPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
}
