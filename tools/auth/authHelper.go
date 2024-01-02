package auth

import (
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"golearn/repository"
)

func UserExists(email string) bool {
	//var count int64
	//utils.DB.Model(&models.User{}).Where("email = ?", email).Count(&count)
	_, err := repository.GetUserByEmail(email)
	if err != nil {
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
