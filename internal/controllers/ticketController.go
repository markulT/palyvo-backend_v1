package controllers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"palyvoua/internal/repository"
	"palyvoua/tools/auth"
	"palyvoua/tools/jsonHelper"
)

type ticketController struct {
	ticketRepo repository.TicketRepo
	userRepo repository.UserRepo
	adminRepo adminRepo
}

//type ticketControllerOptions func(*ticketController)
//
//func WithUserRepo(userRepo userRepository) ticketControllerOptions {
//	return func(controller *ticketController) {
//		controller.userRepo = userRepo
//	}
//}

func SetupTicketRoutes(r *gin.Engine, userRepo repository.UserRepo, tr repository.TicketRepo, adminRepo adminRepo) {
	ticketGroup := r.Group("/ticket")

	tc := ticketController{userRepo: userRepo, ticketRepo: tr, adminRepo: adminRepo}

	ticketGroup.Use(auth.AuthMiddleware(userRepo, adminRepo))
	ticketGroup.Use(auth.RoleMiddleware(0, userRepo, tc.adminRepo))
	ticketGroup.GET("/", jsonHelper.MakeHttpHandler(tc.getAll))
	ticketGroup.GET("/:id", jsonHelper.MakeHttpHandler(tc.getByID))
}

func (tc *ticketController) getAll(c *gin.Context) error {

	authBodyField, exists := c.Get("authBody")
	if !exists {
		return jsonHelper.DefaultHttpErrors["400"]
	}

	authBody, ok := authBodyField.(auth.AuthBody)

	if !ok {
		return jsonHelper.DefaultHttpErrors["400"]
	}
	fmt.Println(authBody.GetUser().ID)
	tickets, err := tc.ticketRepo.GetAllTicketsByUserID(c,authBody.GetUser().ID)
	if err != nil {
		return jsonHelper.ApiError{
			Err:    "Error getting tickets",
			Status: 500,
		}
	}

	c.JSON(200, gin.H{"tickets":tickets})

	return nil
}

func (tc *ticketController) getByID(c *gin.Context) error {

	authBodyField, exists := c.Get("authBody")
	if !exists {
		return jsonHelper.DefaultHttpErrors["400"]
	}

	authBody, ok := authBodyField.(auth.AuthBody)

	if !ok {
		return jsonHelper.DefaultHttpErrors["400"]
	}

	ticketToReceive := c.Param("id")

	ticket, err := tc.ticketRepo.GetByID(uuid.MustParse(ticketToReceive))
	if err != nil {
		return jsonHelper.ApiError{
			Err:    "Error receiving ticket with given id",
			Status: 500,
		}
	}
	if ticket.UserId.String() != authBody.GetUser().ID.String() && authBody.GetRole().AuthorityLevel < 2 {
		return jsonHelper.ApiError{
			Err:    "You have no authority to retrieve this source",
			Status: 403,
		}
	}
	c.JSON(200, gin.H{"ticket":ticket})
	return nil
}
