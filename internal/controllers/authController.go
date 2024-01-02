package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"palyvoua/internal/models"
	"palyvoua/tools/auth"
	"palyvoua/tools/jsonHelper"
)

type authController struct {
	AuthRepo userRepository
	paymentService
}

type userRepository interface {
	SaveUser(*models.User) error
	GetUserByEmail(email string) (models.User, error)
	UpdateCustomerIDByEmail(email string, cid string) error
}

func SetupAuthRoutes(r *gin.Engine, authRepo userRepository) {
	ac := authController{AuthRepo: authRepo}

	authGroup := r.Group("/auth")

	authGroup.POST("/register", jsonHelper.MakeHttpHandler(ac.register))
	authGroup.POST("/login", jsonHelper.MakeHttpHandler(ac.login))
	authGroup.POST("/refresh", jsonHelper.MakeHttpHandler(ac.refresh))
}


type RefreshRequest struct {
	RefreshToken string `json:"refreshToken"`
}

func (ac *authController) refresh(c *gin.Context) error {

	var body RefreshRequest
	if err:=c.Bind(&body);err!=nil{
		return jsonHelper.DefaultHttpErrors["BadRequest"]
	}
	var userFromDb models.User
	email, err := auth.GetSubject(body.RefreshToken)
	if err != nil {
		return jsonHelper.ApiError{
			Err:    err.Error(),
			Status: 400,
		}
	}
	userFromDb, err = ac.AuthRepo.GetUserByEmail(email)
	if err != nil {
		return jsonHelper.ApiError{
			Err:    err.Error(),
			Status: 400,
		}
	}
	if _, err := auth.Validate(body.RefreshToken); err != nil {
		return jsonHelper.ApiError{
			Err:    err.Error(),
			Status: 400,
		}
	}
	tokens := auth.GenerateTokens(map[string]interface{}{
		"email": userFromDb.Email,
	}, c)
	c.JSON(200, gin.H{
		"accessToken":  tokens.AccessToken,
		"refreshToken": tokens.RefreshToken,
	})
	return nil
}


type LoginRequest struct {
	Email string `json:"email"`
	Password string `json:"password"`
}

func (ac *authController) login(c *gin.Context) error {

	var body LoginRequest
	if err:=c.Bind(&body);err!=nil{
		return jsonHelper.DefaultHttpErrors["BadRequest"]
	}

	userFromDB, err := ac.AuthRepo.GetUserByEmail(body.Email)
	if err != nil {
		return jsonHelper.ApiError{
			Err:    err.Error(),
			Status: 404,
		}
	}

	if err := bcrypt.CompareHashAndPassword([]byte(userFromDB.Password), []byte(body.Password)); err != nil {
		return jsonHelper.ApiError{
			Err:    "Invalid password",
			Status: 403,
		}
	}
	tokens := auth.GenerateTokens(map[string]interface{}{
		"email":    userFromDB.Email,
	}, c)

	c.JSON(200, gin.H{
		"accessToken":  tokens.AccessToken,
		"refreshToken": tokens.RefreshToken,
	})
	return nil
}


type RegisterRequest struct {
	Email string `json:"email"`
	Password string `json:"password"`
}

type RegisterResponse struct {
	AccessToken string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

func (ac *authController) register(c *gin.Context) error {

	var body RegisterRequest
	if err := c.Bind(&body);err!=nil {
		return jsonHelper.DefaultHttpErrors["BadRequest"]
	}

	userExists := auth.UserExists(body.Email)
	if userExists {
		return jsonHelper.ApiError{
			Err:    "User already exists",
			Status: 400,
		}
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		return jsonHelper.ApiError{
			Err:    err.Error(),
			Status: 500,
		}
	}
	userId, err := uuid.NewRandom()
	if err != nil {
		return jsonHelper.ApiError{
			Err:    err.Error(),
			Status: 500,
		}
	}
	newUser := models.User{ID: userId, Email: body.Email, Password: string(hashedPassword)}
	tokens := auth.GenerateTokens(map[string]interface{}{
		"email": newUser.Email,
	}, c)

	if err := ac.AuthRepo.SaveUser(&newUser); err != nil {
		return jsonHelper.ApiError{
			Err:    err.Error(),
			Status: 500,
		}
	}

	customerID, err := ac.paymentService.CreateCustomer(body.Email)
	if err != nil {
		return jsonHelper.ApiError{
			Err:    err.Error(),
			Status: 500,
		}
	}
	err = ac.AuthRepo.UpdateCustomerIDByEmail(body.Email, customerID)
	if err != nil {
		return jsonHelper.ApiError{
			Err:    err.Error(),
			Status: 500,
		}
	}

	c.JSON(200, gin.H{
		"accessToken":  tokens.AccessToken,
		"refreshToken": tokens.RefreshToken,
	})
	return nil
}


