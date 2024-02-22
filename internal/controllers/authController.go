package controllers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"palyvoua/internal/models"
	"palyvoua/internal/repository"
	"palyvoua/tools/auth"
	"palyvoua/tools/jsonHelper"
)

type authController struct {
	AuthRepo repository.UserRepo
	paymentService paymentService
}

func SetupAuthRoutes(r *gin.Engine, authRepo repository.UserRepo, ps paymentService) {
	ac := authController{AuthRepo: authRepo, paymentService:ps}

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
	userFromDb, err = ac.AuthRepo.GetUserByEmail(c,email)
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
		"_id":userFromDb.ID.String(),
		"roleId":userFromDb.Role.String(),
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

	userFromDB, err := ac.AuthRepo.GetUserByEmail(c,body.Email)
	if err != nil {
		return jsonHelper.ApiError{
			Err:    err.Error(),
			Status: 404,
		}
	}

	if err := bcrypt.CompareHashAndPassword([]byte(userFromDB.Password), []byte(body.Password)); err != nil {
		fmt.Println(err.Error())
		return jsonHelper.ApiError{
			Err:    "Invalid password",
			Status: 403,
		}
	}
	tokens := auth.GenerateTokens(map[string]interface{}{
		"email": userFromDB.Email,
		"_id":userFromDB.ID.String(),
		"roleId":userFromDB.Role.String(),
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
	adminRepo := repository.NewAdminRepo()
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

	role,err := adminRepo.GetRoleByName("ROLE_ADMIN")
	//role,err := adminRepo.GetRoleByName("ROLE_OPERATOR")
	//role,err := adminRepo.GetRoleByName("ROLE_USER")
	if err != nil {
		return jsonHelper.ApiError{
			Err:    "No role found",
			Status: 404,
		}
	}
	newUser := models.User{
		ID: userId, Email: body.Email, Password: string(hashedPassword), Role:role.ID,
	}
	tokens := auth.GenerateTokens(map[string]interface{}{
		"email": newUser.Email,
		"_id":newUser.ID.String(),
		"roleId":newUser.Role.String(),
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


