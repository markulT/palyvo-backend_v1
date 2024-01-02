package tools

import (
	"github.com/stripe/stripe-go/v75"
	"os"
)

func StripeInit() {
	stripeSecretKey := os.Getenv("stripeSecretKey")
	stripe.Key = stripeSecretKey
}
