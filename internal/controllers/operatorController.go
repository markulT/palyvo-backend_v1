package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"palyvoua/internal/models"
	"palyvoua/internal/repository"
	"palyvoua/tools/auth"
	"palyvoua/tools/jsonHelper"
)

//type operatorControllerOptions func(*operatorController)

type operatorController struct {
	authRepo repository.UserRepo
	adminRepo adminRepo
	ticketRepo repository.TicketRepo
}

func SetupOperatorRoutes(r *gin.Engine , authRepo repository.UserRepo, adminRepo adminRepo, tr repository.TicketRepo) {
	operatorGroup := r.Group("/operator")

	oc := operatorController{authRepo: authRepo, adminRepo: adminRepo, ticketRepo: tr}


	operatorGroup.Use(auth.AuthMiddleware(oc.authRepo, adminRepo))
	operatorGroup.Use(auth.RoleMiddleware(2, oc.authRepo, oc.adminRepo))
	operatorGroup.POST("/submitTicket", jsonHelper.MakeHttpHandler(oc.submitTicket))

}

type SubmitTicketRequest struct {
	TicketID string `json:"ticketId"`
}

func (oc *operatorController) submitTicket(c *gin.Context) error {
	var body SubmitTicketRequest
	if err:=c.Bind(&body);err!=nil {
		return jsonHelper.DefaultHttpErrors["BadRequest"]
	}
	//user, exists := c.Get("user")
	//if !exists {
	//	return jsonHelper.DefaultHttpErrors["BadRequest"]
	//}
	//authorityLevel, exists := c.Get("authorityLevel")
	//if !exists {
	//	return jsonHelper.DefaultHttpErrors["BadRequest"]
	//}

	ticket, err := oc.ticketRepo.GetByID(uuid.MustParse(body.TicketID))
	if err != nil {
		return jsonHelper.ApiError{
			Err:    err.Error(),
			Status: 500,
		}
	}

	if ticket.PaymentID == "" {
		return jsonHelper.ApiError{
			Err:    "Invalid ticket",
			Status: 500,
		}
	}

	err = oc.ticketRepo.UpdateStatus(ticket.ID, models.USED)
	if err != nil {
		return jsonHelper.ApiError{
			Err:    "Error changing ticket's status",
			Status: 0,
		}
	}

	c.JSON(200, gin.H{})
	return nil
}
