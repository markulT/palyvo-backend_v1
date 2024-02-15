package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v75"
	"github.com/stripe/stripe-go/v75/checkout/session"
	"github.com/stripe/stripe-go/v75/paymentmethod"
	"github.com/stripe/stripe-go/v75/webhook"
	"io"
	"os"
	"palyvoua/internal/api/payment"
	"palyvoua/internal/models"
	"palyvoua/internal/repository"
	"palyvoua/tools/auth"
	"palyvoua/tools/jsonHelper"
	"strconv"
	"sync"
	"time"
)

type paymentController struct {
	userRepo userRepository
	paymentService
	ticketRepo repository.TicketRepo
	productRepo repository.ProductRepo
	productTicketRepo repository.ProductTicketRepo
}

type paymentService interface {
	SetDefaultPaymentMethod(customerID string, paymentMethodID string) error
	CreateCustomer(email string) (string, error)
	GetDefaultPaymentMethod(customerID string) (string, error)
	DeletePaymentMethodByIDAndCustomerID(paymentMethodID string, customerID string) error
	ChargeCustomer(customerID string, amount int) (string, error)
	CreateSetupIntent(cid string) (*stripe.SetupIntent, error)
	GetCustomerByID(cid string) (*stripe.Customer, error)
	SaveProduct(product *models.ProductTicket) (*stripe.Product, error)
	CreateCheckoutSession(productList []payment.ProductDto, customerID string) (*stripe.CheckoutSession, error)
}

type PaymentRouterOptions struct {
	UserRepository userRepository
	Ps paymentService
	Tr repository.TicketRepo
	Pr repository.ProductRepo
	Ptr repository.ProductTicketRepo

}

func SetupPaymentRoutes(r *gin.Engine, options *PaymentRouterOptions) {	paymentGroup := r.Group("/payment")
	pc := paymentController{userRepo: options.UserRepository, paymentService: options.Ps, ticketRepo: options.Tr, productRepo: options.Pr, productTicketRepo: options.Ptr}

	paymentGroup.POST("/webhook", jsonHelper.MakeHttpHandler(pc.webhookHandler))

	paymentGroup.Use(auth.AuthMiddleware(options.UserRepository))
	paymentGroup.POST("/method/setDefault", jsonHelper.MakeHttpHandler(pc.setDefaultPaymentMethod))
	paymentGroup.GET("/paymentMethod/getAll", jsonHelper.MakeHttpHandler(pc.paymentMethodsHandler))
	paymentGroup.GET("/paymentMethod/getDefault", jsonHelper.MakeHttpHandler(pc.getDefaultPaymentMethod))
	paymentGroup.DELETE("/paymentMethod/delete/:id", jsonHelper.MakeHttpHandler(pc.deletePaymentMethod))

	//paymentGroup.POST("/buy/amount", jsonHelper.MakeHttpHandler(pc.buyAmount))
	paymentGroup.POST("/setupIntent/create",jsonHelper.MakeHttpHandler(pc.createSetupIntent))
	paymentGroup.POST("/checkout/create",jsonHelper.MakeHttpHandler(pc.createCheckoutSession))
}

func (sc *paymentController) processTicket(c context.Context, wg *sync.WaitGroup, errorCh chan error, dto *payment.ProductDto, user *models.User, sess *stripe.CheckoutSession) {
	defer wg.Done()
	expirationTerm, err := strconv.Atoi(os.Getenv("TICKET_EXPIRATION"))
	if err != nil {
		errorCh <- err
		return
	}

	productTicket,err := sc.productTicketRepo.GetByStripeProductID(c, dto.ProductStripeID)
	if err != nil {
		errorCh <- err
		return
	}
	ticketID, _ := uuid.NewRandom()
	ticket := models.Ticket{
		CreatedAt: int(time.Now().Unix()),
		ExpiresAt: int(time.Now().Add(time.Hour * 24 * time.Duration(expirationTerm)).Unix()),
		ID: ticketID,
		UserId: user.ID,
		Status: models.NOT_ACTIVATED,
		Quantity: dto.Amount,
		ProductTicketID: productTicket.ProductID,
	}
	ticket.SetSecret("Huy")

	err = sc.ticketRepo.Create(c,ticket)
	if err != nil {
		errorCh <- err
		return
	}
	err = sc.ticketRepo.UpdatePaymentID(c,ticketID, sess.PaymentIntent.ID)
	if err != nil {
		errorCh <- err
		return
	}
	err = sc.productRepo.DecreaseProductAmount(c, productTicket.ProductID, productTicket.Amount)
	if err != nil {
		errorCh <- err
		return
	}
	return
}

func (sc *paymentController) webhookHandler(c *gin.Context) error {
	requestBody, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return jsonHelper.ApiError{
			Err:    err.Error(),
			Status: 400,
		}
	}
	event, err := webhook.ConstructEventWithOptions(requestBody, c.GetHeader("Stripe-Signature"), os.Getenv("WEBHOOK_SECRET_STRIPE"), webhook.ConstructEventOptions{IgnoreAPIVersionMismatch: true})
	if err != nil {
		return jsonHelper.ApiError{
			Err:    err.Error(),
			Status: 400,
		}
	}

	switch event.Type {
	case "checkout.session.async_payment_succeeded":

	case "checkout.session.completed":
		var checkoutSession stripe.CheckoutSession
		err:=json.Unmarshal(event.Data.Raw, &checkoutSession)
		if err != nil {
			return jsonHelper.ApiError{
				Err:    err.Error(),
				Status: 500,
			}
		}

		var param ="line_items"
		sess, err := session.Get(event.Data.Object["id"].(string), &stripe.CheckoutSessionParams{
			Expand: []*string{&param},
		})

		user,err:=sc.userRepo.GetByCustomerID(checkoutSession.Customer.ID)
		if err != nil {
			return jsonHelper.ApiError{
				Err:    err.Error(),
				Status: 500,
			}
		}

		err = sc.ticketRepo.WithTransaction(c, func(c context.Context) error {
			var productDtoList []*payment.ProductDto
			wg := sync.WaitGroup{}

			errorCh := make(chan error, 1)

			for _, stripeProduct := range sess.LineItems.Data {
				var dto = payment.ProductDto{
					ProductStripeID: stripeProduct.Price.Product.ID,
					Amount:          int(stripeProduct.Quantity),
				}
				productDtoList = append(productDtoList, &dto)
				wg.Add(1)
				go sc.processTicket(c, &wg,errorCh, &dto, &user, sess)
			}
			select {
			case err = <-errorCh:
				if err != nil {
					return err
				}
			}
			wg.Wait()
			return nil
		})

		if err != nil {
			fmt.Println("chmonya tyt")
			return jsonHelper.ApiError{

				Err:    err.Error(),
				Status: 500,
			}
		}


	}
	return nil
}

type CreateCheckoutSessionRequest struct {
	ProductList []payment.ProductDto `json:"productList" bson:"productList"`
}

func (sc *paymentController) createCheckoutSession(c *gin.Context) error {

	userField, exists := c.Get("user")
	if !exists {
		return jsonHelper.DefaultHttpErrors["400"]
	}
	user, ok := userField.(models.User)
	if !ok {

		return jsonHelper.DefaultHttpErrors["400"]
	}

	var body CreateCheckoutSessionRequest
	if err:=c.Bind(&body);err!=nil{
		return jsonHelper.DefaultHttpErrors["BadRequest"]
	}

	fmt.Println(body)

	sess,err := sc.paymentService.CreateCheckoutSession(body.ProductList, user.CustomerID)
	if err != nil {
		return jsonHelper.ApiError{
			Err:    "Internal server error",
			Status: 500,
		}
	}
	fmt.Println(*sess)
	fmt.Println(sess.PaymentIntent)
	c.JSON(200, gin.H{"sessionId":sess.ID})
	return nil
}

type CreateSetupIntentRequest struct{}

func (sc *paymentController) createSetupIntent(c *gin.Context) error {

	var body CreateSetupIntentRequest
	jsonHelper.BindWithException(&body, c)

	userEmail, exists := c.Get("userEmail")
	if !exists {
		return jsonHelper.ApiError{
			Err:    "Unauthorized",
			Status: 417,
		}
	}
	user, err := sc.userRepo.GetUserByEmail(fmt.Sprintf("%s", userEmail))
	if err != nil {
		return jsonHelper.ApiError{
			Err:    err.Error(),
			Status: 500,
		}
	}

	si, err := sc.paymentService.CreateSetupIntent(user.CustomerID)
	if err != nil {
		return jsonHelper.ApiError{
			Err:    err.Error(),
			Status: 500,
		}
	}
	customer, err := sc.paymentService.GetCustomerByID(user.CustomerID)
	if err != nil {
		return jsonHelper.ApiError{
			Err:    err.Error(),
			Status: 500,
		}
	}
	c.JSON(200, gin.H{
		"setupClientSecret": si.ClientSecret,
		"customerID":        customer.ID,
	})
	return nil
}

type BuyAmountRequest struct {
	Amount int `json:"amount"`
	ProductID string `json:"productID"`
}

func (pc *paymentController) buyAmount(c *gin.Context) error {

	var body BuyAmountRequest
	if err:=c.Bind(&body);err!=nil{
		return jsonHelper.DefaultHttpErrors["BadRequest"]
	}


	userEmail, exists := c.Get("userEmail")
	if !exists {
		return jsonHelper.ApiError{
			Err:    "User does not exist",
			Status: 404,
		}
	}

	user, err := pc.userRepo.GetUserByEmail(userEmail.(string))
	if err != nil {
		return jsonHelper.ApiError{
			Err:    "No such user",
			Status: 500,
		}
	}

	expirationTerm, err := strconv.Atoi(os.Getenv("TICKET_EXPIRATION"))
	if err != nil {
		return jsonHelper.ApiError{
			Err:    "Internal server error",
			Status: 500,
		}
	}
	err = pc.ticketRepo.WithTransaction(c, func(c context.Context) error {
		product,err := pc.productRepo.GetProduct(c, uuid.MustParse(body.ProductID))
		if err != nil {
			return jsonHelper.ApiError{
				Err:    "Internal server error",
				Status: 500,
			}
		}

		ticketID, _ := uuid.NewRandom()
		ticket := models.Ticket{
			CreatedAt: int(time.Now().Unix()),
			ExpiresAt: int(time.Now().Add(time.Hour * 24 * time.Duration(expirationTerm)).Unix()),
			ID: ticketID,
			UserId: user.ID,
			Status: models.NOT_ACTIVATED,
			Amount: body.Amount,
		}
		ticket.SetSecret("Huy")

		err = pc.ticketRepo.Create(c,ticket)
		if err != nil {
			return jsonHelper.ApiError{
				Err:    err.Error(),
				Status: 500,
			}
		}

		pmID, err := pc.paymentService.ChargeCustomer(user.CustomerID, body.Amount*product.Amount)
		if err != nil {

			return jsonHelper.ApiError{
				Err:    err.Error(),
				Status: 500,
			}
		}
		pc.ticketRepo.UpdatePaymentID(c,ticketID, pmID)
		err = pc.productRepo.DecreaseProductAmount(c, product.ID, body.Amount)
		if err != nil {
			return jsonHelper.ApiError{
				Err:    err.Error(),
				Status: 500,
			}
		}
		return nil
	})

	if err != nil {
		return jsonHelper.ApiError{
			Err:    err.Error(),
			Status: 500,
		}
	}

	return nil
}
type SetDefaultPaymentMethodRequest struct {
	PaymentMethodID string `json:"paymentMethodId"`
}

func (pc *paymentController) setDefaultPaymentMethod(c *gin.Context) error {
	var err error
	var body SetDefaultPaymentMethodRequest
	if err:=c.Bind(&body);err!=nil {
		return jsonHelper.DefaultHttpErrors["BadRequest"]
	}
	authUserEmail, exists := c.Get("userEmail")
	if !exists {
		return jsonHelper.ApiError{
			Err:    "User unauthorized",
			Status: 417,
		}
	}
	user, err := pc.userRepo.GetUserByEmail(fmt.Sprintf("%s", authUserEmail))
	if err != nil {
		return jsonHelper.ApiError{
			Err:    err.Error(),
			Status: 417,
		}
	}
	err = pc.paymentService.SetDefaultPaymentMethod(user.CustomerID, body.PaymentMethodID)
	c.JSON(200, gin.H{})
	return nil
}

func (pc *paymentController) paymentMethodsHandler(c *gin.Context) error {
	userEmail, exists := c.Get("userEmail")
	if !exists {
		return jsonHelper.ApiError{
			Err:    "User does not exist",
			Status: 404,
		}
	}
	user, err := pc.userRepo.GetUserByEmail(fmt.Sprintf("%s", userEmail))
	if err != nil {
		return jsonHelper.ApiError{
			Err:    err.Error(),
			Status: 500,
		}
	}
	var paymentMethods []stripe.PaymentMethod

	params := &stripe.PaymentMethodListParams{
		Customer: stripe.String(user.CustomerID),
		Type:     stripe.String("card"),
	}
	i := paymentmethod.List(params)
	for i.Next() {
		pm := i.PaymentMethod()
		paymentMethods = append(paymentMethods, *pm)
	}

	c.JSON(200, gin.H{"paymentMethods": paymentMethods})
	return nil
}

func (pc *paymentController) getDefaultPaymentMethod(c *gin.Context) error {

	var err error

	authUserEmail, exists := c.Get("userEmail")
	if !exists {
		return jsonHelper.ApiError{
			Err:    "User unauthorized",
			Status: 417,
		}
	}
	user, err := pc.userRepo.GetUserByEmail(fmt.Sprintf("%s", authUserEmail))
	if err != nil {
		return jsonHelper.ApiError{
			Err:    err.Error(),
			Status: 417,
		}
	}

	paymentMethodID, err := pc.paymentService.GetDefaultPaymentMethod(user.CustomerID)
	if err != nil {
		return jsonHelper.ApiError{
			Err:    err.Error(),
			Status: 500,
		}
	}

	c.JSON(200, gin.H{"paymentMethodId":paymentMethodID})
	return nil
}

func (pc *paymentController) deletePaymentMethod(c *gin.Context) error {
	var err error

	authUserEmail, exists := c.Get("userEmail")
	if !exists {
		return jsonHelper.ApiError{
			Err:    "User unauthorized",
			Status: 417,
		}
	}
	user, err := pc.userRepo.GetUserByEmail(fmt.Sprintf("%s", authUserEmail))
	if err != nil {
		return jsonHelper.ApiError{
			Err:    err.Error(),
			Status: 417,
		}
	}

	paymentMethodID := c.Param("id")
	err = pc.paymentService.DeletePaymentMethodByIDAndCustomerID(paymentMethodID,user.CustomerID)
	if err != nil {
		return jsonHelper.ApiError{
			Err:    "Error deleting the payment method",
			Status: 500,
		}
	}

	c.JSON(200, gin.H{})
	return nil
}