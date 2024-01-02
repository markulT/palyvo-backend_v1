package main

import (
	"github.com/gin-gonic/gin"
	"os"
	"palyvoua/internal/controllers"
	"palyvoua/tools"
)

func init() {
	tools.LoadEnvVariables()
	tools.ConnectToDb()
	tools.ConnectToPostgres()
}

func main() {

	r := gin.Default()

	controllers.SetupAuthRoutes(r, nil)

	r.Run(os.Getenv("PORT"))

}