package controllers

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"palyvoua/internal/models"
	"palyvoua/internal/repository"
	"palyvoua/tools/jsonHelper"
)

type productController struct {
	productRepo repository.ProductRepo
}

func SetupProductRoutes(r *gin.Engine, pr repository.ProductRepo, ur userRepository, adminRepo adminRepo) {
	productGroup := r.Group("/product")
	pc := productController{productRepo: pr}

	//productGroup.Use(auth.AuthMiddleware(ur))
	//productGroup.POST("/buy", jsonHelper.MakeHttpHandler(pc.buyProduct))
	productGroup.GET("/all", jsonHelper.MakeHttpHandler(pc.getAllProducts))
	productGroup.GET("/:id", jsonHelper.MakeHttpHandler(pc.createProduct))
	//productGroup.Use(auth.RoleMiddleware(3, ur, adminRepo))
	productGroup.POST("/", jsonHelper.MakeHttpHandler(pc.createProduct))
	productGroup.POST("/updateAmount", jsonHelper.MakeHttpHandler(pc.createProduct))
}

func (pc *paymentController) getProductByID(c *gin.Context) error {
	productIDField := c.Param("id")
	productID, err := uuid.Parse(productIDField)
	if err != nil {
		return jsonHelper.DefaultHttpErrors["BadRequest"]
	}

	product, err := pc.productRepo.GetProduct(context.Background(), productID)
	if err != nil {
		return jsonHelper.DefaultHttpErrors["InternalServerError"]
	}
	c.JSON(200, gin.H{"product":product})
	return nil
}

type CreateProductRequest struct {
	Amount int `json:"amount" bson:"amount"`
	Title string `json:"title" bson:"title"`
	Price int `json:"price" bson:"price"`
	Currency string `json:"currency" bson:"currency"`
}

func (pc *productController) getAllProducts(c *gin.Context) error {

	var err error

	products, err := pc.productRepo.GetAllProducts(context.Background())
	if err != nil {
		return jsonHelper.ApiError{
			Err:    "No products",
			Status: 500,
		}
	}
	c.JSON(200, gin.H{"products":products})
	return nil
}

func (pc *productController) createProduct(c *gin.Context) error {
	fmt.Println("creating...")
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
	err = pc.productRepo.SaveProduct(context.Background(), &p)
	if err != nil {
		fmt.Println("error")
		fmt.Println(err.Error())
		return jsonHelper.ApiError{
			Err:    "Internal server error",
			Status: 500,
		}
	}
	c.JSON(200, gin.H{})
	return nil
}

type UpdateProductAmountRequest struct {
	ProductID string `json:"productId"`
	Amount int `json:"amount"`
}

func (pc *productController) updateProductAmount(c *gin.Context) error {
	var err error
	var body UpdateProductAmountRequest
	if err:=c.Bind(&body);err!=nil {
		return jsonHelper.DefaultHttpErrors["BadRequest"]
	}
	pid, err := uuid.Parse(body.ProductID)
	if err != nil {
		return jsonHelper.DefaultHttpErrors["BadRequest"]
	}
	err = pc.productRepo.UpdateProductAmount(context.Background(), pid , body.Amount)
	if err != nil {
		return jsonHelper.ApiError{
			Err:    "Error updating product amount",
			Status: 500,
		}
	}
	c.JSON(200, gin.H{})
	return nil
}