package main

import (
	"github.com/gin-contrib/cors"
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
	controllers.SetupProductRoutes(r, consistentProductRepo, userRepo, adminRepo)
	controllers.SetupTicketRoutes(r,userRepo, ticketRepo,adminRepo)

	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:4200"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	config.AllowCredentials = true
	config.AddAllowHeaders("Access-Control-Allow-Credentials")

	r.Use(cors.New(config))


	r.Run()

}