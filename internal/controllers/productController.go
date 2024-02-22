package controllers

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"palyvoua/internal/models"
	"palyvoua/internal/repository"
	"palyvoua/tools/auth"
	"palyvoua/tools/jsonHelper"
)

type productController struct {
	productRepo repository.ProductRepo
	paymentService paymentService
}

func SetupProductRoutes(r *gin.Engine, pr repository.ProductRepo, ur repository.UserRepo, adminRepo adminRepo, ps paymentService) {
	productGroup := r.Group("/product")
	pc := productController{productRepo: pr, paymentService: ps}

	productGroup.Use(auth.AuthMiddleware(ur, adminRepo))
	//productGroup.POST("/buy", jsonHelper.MakeHttpHandler(pc.buyProduct))
	productGroup.GET("/all", jsonHelper.MakeHttpHandler(pc.getAllProducts))
	productGroup.GET("/:id", jsonHelper.MakeHttpHandler(pc.getProductByID))
	productGroup.GET("/byFuelType", jsonHelper.MakeHttpHandler(pc.getByFuelType))
	productGroup.GET("/bySeller", jsonHelper.MakeHttpHandler(pc.getBySeller))
	productGroup.Use(auth.RoleMiddleware(3, ur, adminRepo))
	productGroup.POST("/", jsonHelper.MakeHttpHandler(pc.createProduct))
	productGroup.POST("/updateAmount", jsonHelper.MakeHttpHandler(pc.updateProductAmount))
	productGroup.DELETE("/delete", jsonHelper.MakeHttpHandler(pc.updateProductAmount))
}

func (pc *productController) deleteProduct(c *gin.Context) error {
	var err error
	productIDField := c.Query("productId")

	productID, err := uuid.Parse(productIDField)
	if err !=nil {
		return jsonHelper.ApiError{
			Err:    "Error parsing id",
			Status: 500,
		}
	}

	err = pc.productRepo.DeleteProduct(c, productID)
	if err !=nil {
		return jsonHelper.ApiError{
			Err:    "Error deleting product: " + err.Error(),
			Status: 500,
		}
	}

	c.JSON(200, gin.H{})

	return nil
}

type CreateStripeProductRequest struct {
	Price int `json:"price"`

}

func (pc *productController) getByFuelType(c *gin.Context) error {
	fuelTypeField := c.Query("fuelType")
	var products []models.Product
	var err error

	products, err = pc.productRepo.GetByFuelType(context.Background(), fuelTypeField)
	if err != nil {
		return jsonHelper.ApiError{
			Err:    "Error occurred while receiving product",
			Status: 500,
		}
	}

	c.JSON(200, gin.H{"products":products})
	return nil
}


func (pc *productController) getBySeller(c *gin.Context) error {
	fuelTypeField := c.Query("seller")
	var products []models.Product
	var err error

	products, err = pc.productRepo.GetBySeller(context.Background(), fuelTypeField)
	if err != nil {
		return jsonHelper.ApiError{
			Err:    "Error occurred while receiving product",
			Status: 500,
		}
	}

	c.JSON(200, gin.H{"products":products})
	return nil
}

func (pc *productController) getProductByID(c *gin.Context) error {
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
	Seller string `json:"seller" bson:"seller"`
	FuelType string `json:"fuelType" bson:"fuelType"`
}

func (pc *productController) getAllProducts(c *gin.Context) error {

	var err error

	products, err := pc.productRepo.GetAllProducts(context.Background())
	if err != nil {
		fmt.Println(err.Error())
		return jsonHelper.ApiError{
			Err:    "No products",
			Status: 500,
		}
	}
	c.JSON(200, gin.H{"products":products})
	return nil
}

func (pc *productController) createProduct(c *gin.Context) error {
	fmt.Println("at least came here")
	var body CreateProductRequest
	if err := c.Bind(&body);err!=nil{
		fmt.Println(err)
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
		Seller: body.Seller,
		FuelType: body.FuelType,
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

	if err != nil {
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