package payment

import (
	"fmt"
	"github.com/stripe/stripe-go/v75"
	"github.com/stripe/stripe-go/v75/charge"
	"github.com/stripe/stripe-go/v75/checkout/session"
	"github.com/stripe/stripe-go/v75/customer"
	"github.com/stripe/stripe-go/v75/paymentmethod"
	"github.com/stripe/stripe-go/v75/price"
	"github.com/stripe/stripe-go/v75/product"
	"github.com/stripe/stripe-go/v75/setupintent"
	"palyvoua/internal/models"
)

type PaymentError struct {}
func (e PaymentError) Error() string {
	return "Generic error"
}

type PaymentService interface {
	SetDefaultPaymentMethod(customerID string, paymentMethodID string) error
	CreateCustomer(email string) (string, error)
	GetDefaultPaymentMethod(customerID string) (string, error)
	DeletePaymentMethodByIDAndCustomerID(paymentMethodID string, customerID string) error
	ChargeCustomer(customerID string, amount int) (string, error)
	CreateSetupIntent(cid string) (*stripe.SetupIntent, error)
	GetCustomerByID(cid string) (*stripe.Customer, error)
	SaveProduct(p *models.ProductTicket) (*stripe.Product,error)
	CreateCheckoutSession(productList []ProductDto, customerID string) (*stripe.CheckoutSession, error)
}

type ProductDto struct {
	ProductStripeID string `bson:"productStripeId" json:"productStripeId"`
	Amount int `json:"amount" bson:"amount"`
}

func NewStripePaymentService() PaymentService {
	return &stripePaymentService{}
}

type stripePaymentService struct {

}

func (s *stripePaymentService) CreateCheckoutSession(productList []ProductDto, customerID string) (*stripe.CheckoutSession, error) {

	var lineItems []*stripe.CheckoutSessionLineItemParams

	for _, p := range productList {

		quanity := int64(p.Amount)
		lineItem := stripe.CheckoutSessionLineItemParams{
			Price: stripe.String(p.ProductStripeID),
			Quantity: &quanity,
		}
		lineItems = append(lineItems, &lineItem)
	}
	//i := price.List(priceParams)
	//var priceId string
	//for i.Next() {
	//	p := i.Price()
	//	priceId = p.ID
	//}

	// Create a checkout session with the product's price
	params := &stripe.CheckoutSessionParams{
		Customer: stripe.String(customerID),
		PaymentMethodTypes: stripe.StringSlice([]string{
			"card",
		}),
		LineItems: lineItems,
		Mode: stripe.String(string(stripe.CheckoutSessionModePayment)),
		SuccessURL: stripe.String("https://example.com/success"),
		CancelURL: stripe.String("https://example.com/cancel"),
	}

	sess, err := session.New(params)
	return sess,err
}


func (s *stripePaymentService) SaveProduct(p *models.ProductTicket) (*stripe.Product,error) {
	stripeProduct, err := product.New(&stripe.ProductParams{
		Name: stripe.String(p.Title),
		Type: stripe.String(string(stripe.ProductTypeGood)),
		Metadata: map[string]string{
			"title":     p.Title,
			"productID": p.ID.String(),
		},
	})
	if err != nil {
		return nil, err
	}
	_, err = price.New(&stripe.PriceParams{
		Product:    stripe.String(stripeProduct.ID),
		UnitAmount: stripe.Int64(int64(p.Price)), // price in cents
		Currency:   stripe.String(p.Currency),
	})
	if err != nil {
		return nil, err
	}

	return stripeProduct,nil
}

func (s *stripePaymentService) GetCustomerByID(cid string) (*stripe.Customer, error) {
	c, err := customer.Get(cid, nil)
	if err != nil {
		return &stripe.Customer{}, err
	}
	return c, nil
}

func (s *stripePaymentService) CreateSetupIntent(cid string) (*stripe.SetupIntent, error) {
	params := &stripe.SetupIntentParams{
		AutomaticPaymentMethods: &stripe.SetupIntentAutomaticPaymentMethodsParams{
			Enabled: stripe.Bool(true),
		},
		Customer: stripe.String(cid),
	}
	si, err := setupintent.New(params)
	if err != nil {
		return &stripe.SetupIntent{}, err
	}
	return si, nil
}

func (s *stripePaymentService) SetDefaultPaymentMethod(customerID string, paymentMethodID string) error {
	params := &stripe.CustomerParams{
		InvoiceSettings: &stripe.CustomerInvoiceSettingsParams{
			DefaultPaymentMethod: stripe.String(customerID),
		},
	}

	_, err := customer.Update(customerID, params)
	return err
}

func (s *stripePaymentService) CreateCustomer(email string) (string, error) {
	params := &stripe.CustomerParams{
		Email: &email,
	}
	c, err := customer.New(params)
	if err != nil {
		return "", err
	}
	return c.ID, err
}

func (s *stripePaymentService) GetDefaultPaymentMethod(customerID string) (string, error) {
	a := "invoice_settings.default_payment_method"

	params := &stripe.CustomerParams{
		Expand: []*string{&a},
	}

	c, err := customer.Get(customerID, params)
	if err != nil {
		return "", err
	}
	return c.InvoiceSettings.DefaultPaymentMethod.ID, nil
}



func (s *stripePaymentService) DeletePaymentMethodByIDAndCustomerID(paymentMethodID string, customerID string) error {
	pm, err := paymentmethod.Get(paymentMethodID, nil)
	if err != nil {
		return err
	}
	if pm.Customer.ID != customerID {
		return PaymentError{}
	}

	_, err = paymentmethod.Detach(paymentMethodID, nil)
	if err != nil {
		return err
	}
	return nil
}

func (s *stripePaymentService) ChargeCustomer(customerID string, amount int) (string, error) {
	params := &stripe.ChargeParams{
		Amount:   stripe.Int64(int64(amount)),
		Currency: stripe.String(string(stripe.CurrencyUAH)),
		Customer: stripe.String(customerID),
	}
	ch, err := charge.New(params)
	if err != nil {
		fmt.Println("Failed to charge customer:", err)
		return "", err
	}

	return ch.ID, nil
}
