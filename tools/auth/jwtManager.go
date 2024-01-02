package auth

import (
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"golearn/utils"
	"os"
	"time"
)

type EmailReadingError struct {
	Message string
}
func (e EmailReadingError)Error() string {
	return e.Message
}

type Tokens struct {
	AccessToken string
	RefreshToken string
}

func createToken(body map[string]interface{}, expirationTime time.Time, secretKey []byte) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	for key, value := range body {
		claims[key] = value
	}
	claims["exp"] = expirationTime.Unix()
	signedToken, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}
	return signedToken, err
}

func CreateAccessToken(body map[string]interface{}) (string, error) {

	expirationTime := utils.GetEnvInt("refreshExpirationTimeDays", 30)
	expirationTimeUnix := time.Now().Add(time.Duration(expirationTime) * time.Hour)

	return createToken(body, expirationTimeUnix, []byte(os.Getenv("secretKey")))
}
func CreateRefreshToken(body map[string]interface{}) (string, error) {
	expirationTime := utils.GetEnvInt("refreshExpirationTimeDays", 30)
	expirationTimeUnix := time.Now().Add(time.Duration(expirationTime) * 24 * time.Hour)
	return createToken(body,expirationTimeUnix, []byte(os.Getenv("secretKey")))
}

func Validate(tokenString string) (*jwt.Token, error) {
	secretKey := os.Getenv("secretKey")
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})
	if err!=nil {
		return nil,err
	}
	if _,ok := token.Claims.(jwt.Claims);!ok && !token.Valid {
		return nil, fmt.Errorf("Invalid token")
	}
	if token.Claims.(jwt.MapClaims)["exp"].(float64)<float64(time.Now().Unix()) {
		return nil, fmt.Errorf("Expired token is being used")
	}
	return token, nil
}

func GetSubject(tokenString string) (string, error)  {
	secretKey := os.Getenv("secretKey")
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})
	if err!=nil {
		return "" ,err
	}
	//a
	claims:=token.Claims.(jwt.MapClaims)
	return claims["email"].(string), nil
}