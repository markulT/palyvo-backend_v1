package controllers

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"palyvoua/internal/models"
	"palyvoua/internal/repository"
	"palyvoua/tools"
	"palyvoua/tools/auth"
	"palyvoua/tools/jsonHelper"
)

type productTicketController struct {
	paymentService paymentService
	productTicketRepo repository.ProductTicketRepo
}

func SetupProductTicketRoutes(r *gin.Engine, adminRepo adminRepo, ur userRepository, productTicketRepo repository.ProductTicketRepo, ps paymentService) {
	productTicketGroup := r.Group("/productTicket")

	ptc := productTicketController{productTicketRepo: productTicketRepo, paymentService: ps}

	productTicketGroup.Use(auth.AuthMiddleware(ur))

	productTicketGroup.GET("/all", jsonHelper.MakeHttpHandler(ptc.getAllProductTickets))
	productTicketGroup.GET("", jsonHelper.MakeHttpHandler(ptc.getByID))
	productTicketGroup.GET("/params", jsonHelper.MakeHttpHandler(ptc.getByParams))

	productTicketGroup.Use(auth.RoleMiddleware(3, ur, adminRepo))

	productTicketGroup.POST("/create", jsonHelper.MakeHttpHandler(ptc.createProductTicket))
	productTicketGroup.DELETE("", jsonHelper.MakeHttpHandler(ptc.deleteProductTicket))
	productTicketGroup.PUT("", jsonHelper.MakeHttpHandler(ptc.updateProductTicket))

}

type UpdateProductTicketRequest struct {
	CreateProductTicketRequest
	ID string `json:"id"`
}

func (ptc *productTicketController) getByParams(c *gin.Context) error {

	var err error
	operator := c.Query("operator")
	fuelType := c.Query("fuelType")

	params := repository.FindTicketParams{}

	if operator != "" {
		params.Operator = &operator
	}

	if fuelType != "" {
		params.FuelType = &fuelType
	}

	ptList, err := ptc.productTicketRepo.FindByParams(c, &params)

	if err !=nil {
		return jsonHelper.ApiError{
			Err:    "Error getting product ticket by operator",
			Status: 500,
		}
	}



	c.JSON(200, gin.H{"productTickets":ptList})
	return nil
}

func (ptc *productTicketController) getByID(c *gin.Context) error {

	var err error
	id := c.Query("id")
	ptID, err := uuid.Parse(id)
	if err !=nil {
		return jsonHelper.ApiError{
			Err:    "Invalid id type",
			Status: 400,
		}
	}
	pt, err := ptc.productTicketRepo.GetByID(c, ptID)

	if err!= nil {
		return jsonHelper.ApiError{
			Err:    "Error getting product ticket by id: " + err.Error(),
			Status: 500,
		}
	}

	c.JSON(200, gin.H{"productTicket":pt})
	return nil
}

func (ptc *productTicketController) updateProductTicket(c *gin.Context) error {
	var err error
	var body UpdateProductTicketRequest
	if err=c.Bind(&body);err!=nil {
		return jsonHelper.DefaultHttpErrors["BadRequest"]
	}

	productTicketID, err := uuid.Parse(body.ID)
	if err=c.Bind(&body);err!=nil {
		return jsonHelper.DefaultHttpErrors["BadRequest"]
	}

	originalProductTicket, err := ptc.productTicketRepo.GetByID(context.Background(),productTicketID)

	newProductTicket := models.ProductTicket{
		ID:        originalProductTicket.ID,
		ProductID: body.ProductID,
		Amount:    body.Amount,
		Title:     body.Title,
		Price:     body.Price,
		Currency:  body.Currency,
		StripeID:  originalProductTicket.StripeID,
	}

	err = ptc.productTicketRepo.UpdateProductTicket(context.Background(),originalProductTicket.ID, &newProductTicket)
	if err!=nil {
		return jsonHelper.ApiError{
			Err:    "Error updating product ticket",
			Status: 500,
		}
	}
	c.JSON(200,gin.H{})
	return nil
}


type CreateProductTicketRequest struct {
	ProductID uuid.UUID `json:"productId"`
	Amount int `json:"amount"`
	Title string `json:"title"`
	Price int `json:"price"`
	Currency string `json:"currency"`
	Seller string `bson:"seller" json:"seller"`
	FuelType string `json:"fuelType" bson:"fuelType"`
}

func (ptc *productTicketController) deleteProductTicket(c *gin.Context) error {
	var err error
	productTicketIDStr := c.Query("productTicketId")
	productTicketID , err := uuid.Parse(productTicketIDStr)
	if err != nil {
		return jsonHelper.ApiError{
			Err:    "Error parsing id",
			Status: 400,
		}
	}
	_,err = tools.WithTransaction(c, func(ctx context.Context) (interface{}, error) {
		err = ptc.productTicketRepo.DeleteProductTicket(c,productTicketID)
		if err != nil {
			return nil,err
		}
		return nil, nil
	})

	c.JSON(200,gin.H{})
	return nil
}

func (ptc *productTicketController) createProductTicket(c *gin.Context) error {
	var err error
	var body CreateProductTicketRequest
	if err=c.Bind(&body);err!=nil {
		return jsonHelper.DefaultHttpErrors["BadRequest"]
	}
	productTicketID , err := uuid.NewRandom()
	if err != nil {
		return jsonHelper.ApiError{
			Err:    "Error saving generating id",
			Status: 500,
		}
	}

	productTicket := models.ProductTicket{
		ID:        productTicketID,
		ProductID: body.ProductID,
		Amount:    body.Amount,
		Title:     body.Title,
		Price:     body.Price,
		Currency:  body.Currency,
		StripeID:  "",
		Seller: body.Seller,
		FuelType: body.FuelType,
	}

	err = ptc.productTicketRepo.SaveProductTicket(context.Background(),&productTicket)
	if err != nil {
		return jsonHelper.ApiError{
			Err:    "Error saving product ticket",
			Status: 500,
		}
	}

	stripeProduct,err := ptc.paymentService.SaveProduct(&productTicket)
	if err != nil {
		return jsonHelper.ApiError{
			Err:    "Error saving product ticket into stripe",
			Status: 500,
		}
	}
	fmt.Println(stripeProduct.ID)
	updatedProductTicket, err := ptc.productTicketRepo.UpdateStripeProductID(context.Background(),productTicketID, stripeProduct.ID)
	if err != nil {
		return jsonHelper.ApiError{
			Err:    "Error updating product ticket",
			Status: 500,
		}
	}
	c.JSON(200, gin.H{"productTicket":updatedProductTicket})
	return nil
}

func (ptc *productTicketController) getAllProductTickets(c *gin.Context) error {
	var productTickets []models.ProductTicket
	var err error
	productTickets, err = ptc.productTicketRepo.GetAllProductTickets(context.Background())
	if err != nil {
		return jsonHelper.ApiError{
			Err:    "Couldn't receive all product tickets",
			Status: 500,
		}
	}
	c.JSON(200, gin.H{"productTickets":productTickets})
	return nil
}


