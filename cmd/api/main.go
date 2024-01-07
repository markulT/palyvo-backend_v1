package main

import (
	"github.com/gin-gonic/gin"
	"palyvoua/internal/api/payment"
	"palyvoua/internal/controllers"
	"palyvoua/internal/repository"
	"palyvoua/tools"
)

func init() {
	tools.LoadEnvVariables()
	tools.ConnectToDb()
	tools.ConnectToPostgres()
	tools.StripeInit()
}

func main() {

	r := gin.Default()

	userRepo:=repository.NewUserRepo()
	adminRepo := repository.NewAdminRepo()
	ticketRepo := repository.NewTickerRepo()
	stripePaymentService := payment.NewStripePaymentService()
	consistentProductRepo := repository.NewConsistentProductRepo()

	controllers.SetupAuthRoutes(r, userRepo, stripePaymentService)
	controllers.SetupOperatorRoutes(r, userRepo, adminRepo, ticketRepo)
	controllers.SetupPaymentRoutes(r, userRepo,stripePaymentService, ticketRepo, consistentProductRepo)
	controllers.SetupAdminRoutes(r, adminRepo, userRepo)
	controllers.SetupProductRoutes(r, consistentProductRepo)
	controllers.SetupTicketRoutes(r,userRepo, ticketRepo,adminRepo)

	r.Run()

}