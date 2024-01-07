package tools

import (
	"github.com/stripe/stripe-go/v75"
	"os"
)

func StripeInit() {
	stripeSecretKey := os.Getenv("STRIPE_SECRET_KEY")
	stripe.Key = stripeSecretKey
}
