package controllers

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"palyvoua/internal/models"
	"palyvoua/internal/repository"
	"palyvoua/tools/auth"
	"palyvoua/tools/jsonHelper"
)

type productController struct {
	productRepo repository.ProductRepo
}

func SetupProductRoutes(r *gin.Engine, pr repository.ProductRepo, ur userRepository, adminRepo adminRepo) {
	productGroup := r.Group("/product")
	pc := productController{productRepo: pr}

	productGroup.Use(auth.AuthMiddleware(ur))
	productGroup.POST("/buy", jsonHelper.MakeHttpHandler(pc.buyProduct))

	productGroup.Use(auth.RoleMiddleware(3, ur, adminRepo))
	productGroup.POST("/", jsonHelper.MakeHttpHandler(pc.createProduct))
	productGroup.POST("/", jsonHelper.MakeHttpHandler(pc.createProduct))
}

func (pc *productController) buyProduct(c *gin.Context) error {



	return nil
}

type CreateProductRequest struct {
	Amount int `json:"amount" bson:"amount"`
	Title string `json:"title" bson:"title"`
	Price int `json:"price" bson:"price"`
	Currency string `json:"currency" bson:"currency"`
}

func (pc *productController) createProduct(c *gin.Context) error {

	var body CreateProductRequest
	if err := c.Bind(&body);err!=nil {
		return jsonHelper.DefaultHttpErrors["BadRequest"]
	}
	productID,err := uuid.NewRandom()
	if err != nil {
		return jsonHelper.ApiError{
			Err:    err.Error(),
			Status: 500,
		}
	}
	p := models.Product{
		Amount:   body.Amount,
		ID:       productID,
		Title:    body.Title,
		Price:    body.Price,
		Currency: body.Currency,
	}
	pc.productRepo.SaveProduct(context.Background(), &p)
	c.JSON(200, gin.H{})
	return nil
}
