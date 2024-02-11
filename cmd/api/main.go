package main

import (
	"github.com/gin-gonic/gin"
	"log"
	"os"
	"palyvoua/internal/api/payment"
	"palyvoua/internal/controllers"
	"palyvoua/internal/repository"
	"palyvoua/tools"
	"palyvoua/tools/data"
	"palyvoua/tools/jsonHelper"
)

func init() {

	tools.LoadEnvVariables()
	tools.ConnectToDb()
	tools.ConnectToPostgres()
	tools.StripeInit()
	if os.Getenv("SEED_DB") == "true" {
		log.Println("Seeding database...")
		seeder:=data.NewDBSeeder()
		err := seeder.SeedDB()
		if err != nil {
			panic(err)
		}
	}
}

func main() {

	r := gin.Default()

	userRepo:=repository.NewUserRepo()
	adminRepo := repository.NewAdminRepo()
	ticketRepo := repository.NewTickerRepo()
	stripePaymentService := payment.NewStripePaymentService()
	consistentProductRepo := repository.NewConsistentProductRepo()
	productTicketRepo := repository.NewProductTicketRepo()

	r.Use(jsonHelper.CORSMiddleware())

	//config := cors.DefaultConfig()
	//config.AllowOrigins = []string{"http://localhost:4200"}
	//config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	//config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	//config.AllowCredentials = true
	//config.AddAllowHeaders("Access-Control-Allow-Credentials")
	//
	//r.Use(cors.New(config))


	controllers.SetupAuthRoutes(r, userRepo, stripePaymentService)
	controllers.SetupOperatorRoutes(r, userRepo, adminRepo, ticketRepo)
	controllers.SetupPaymentRoutes(r, userRepo,stripePaymentService, ticketRepo, consistentProductRepo, productTicketRepo)
	controllers.SetupAdminRoutes(r, adminRepo, userRepo)
	controllers.SetupProductRoutes(r, consistentProductRepo, userRepo, adminRepo, stripePaymentService)
	controllers.SetupTicketRoutes(r,userRepo, ticketRepo,adminRepo)
	controllers.SetupProductTicketRoutes(r,adminRepo, userRepo, productTicketRepo, stripePaymentService)

	r.Run()

}