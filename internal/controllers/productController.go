package controllers

import (
	"github.com/gin-gonic/gin"
	"palyvoua/internal/repository"
	"palyvoua/tools/jsonHelper"
)

type productController struct {
	productRepo repository.ProductRepo
}

func SetupProductRoutes(r *gin.Engine, pr repository.ProductRepo) {
	productGroup := r.Group("/product")
	pc := productController{productRepo: pr}
	productGroup.POST("/buy", jsonHelper.MakeHttpHandler(pc.buyProduct))
}

func (pc *productController) buyProduct(c *gin.Context) error {

	return nil
}
