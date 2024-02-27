package controllers

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"palyvoua/internal/dto"
	"palyvoua/internal/mapper"
	"palyvoua/internal/models"
	"palyvoua/internal/repository"
	"palyvoua/tools/auth"
	"palyvoua/tools/jsonHelper"
	"sync"
)

type ticketController struct {
	ticketRepo repository.TicketRepo
	userRepo repository.UserRepo
	adminRepo adminRepo
	ticketMapper mapper.TicketMapper
}

//type ticketControllerOptions func(*ticketController)
//
//func WithUserRepo(userRepo userRepository) ticketControllerOptions {
//	return func(controller *ticketController) {
//		controller.userRepo = userRepo
//	}
//}

type TicketRoutesOptions struct {
	UserRepo repository.UserRepo
	TicketRepo repository.TicketRepo
	AdminRepo repository.AdminRepo
	TicketMapper mapper.TicketMapper
}


func SetupTicketRoutes(r *gin.Engine, options *TicketRoutesOptions) {
	ticketGroup := r.Group("/ticket")

	tc := ticketController{userRepo: options.UserRepo, ticketRepo: options.TicketRepo, adminRepo: options.AdminRepo, ticketMapper: options.TicketMapper}

	ticketGroup.Use(auth.AuthMiddleware(options.UserRepo, options.AdminRepo))
	ticketGroup.Use(auth.RoleMiddleware(0, options.UserRepo, tc.adminRepo))
	ticketGroup.GET("/", jsonHelper.MakeHttpHandler(tc.getAll))
	ticketGroup.GET("/:id", jsonHelper.MakeHttpHandler(tc.getByID))
}

func (tc *ticketController) processTicketMapping(c context.Context,wg *sync.WaitGroup, errorCh chan error, ticket models.Ticket, dtoCh chan dto.FullTicketDto ) {
	defer wg.Done()

	fullDto, err := tc.ticketMapper.ModelToFullDto(c,&ticket)
	if err != nil {
		errorCh <- err
		return
	}
	dtoCh <- *fullDto
	return
}

func (tc *ticketController) mapTicketList(c context.Context, ticketList *[]models.Ticket) ([]dto.FullTicketDto, error) {
	var dtoList []dto.FullTicketDto
	wg := sync.WaitGroup{}
	var err error
	errorCh := make(chan error, len(*ticketList))
	respCh := make(chan dto.FullTicketDto, len(*ticketList))
	for _, ticket := range *ticketList {
		wg.Add(1)
		go tc.processTicketMapping(c,&wg, errorCh, ticket,respCh)
	}

	wg.Wait()
	close(errorCh)
	close(respCh)

	for err = range errorCh {
		if err != nil {
			return nil, err
		}
	}

	for fullDto := range respCh {
		dtoList = append(dtoList, fullDto)
	}
	return dtoList, nil
}

func (tc *ticketController) getAll(c *gin.Context) error {
	var ticketDtos []dto.FullTicketDto
	authBodyField, exists := c.Get("authBody")
	if !exists {
		return jsonHelper.DefaultHttpErrors["400"]
	}

	authBody, ok := authBodyField.(auth.AuthBody)

	if !ok {
		return jsonHelper.DefaultHttpErrors["400"]
	}
	tickets, err := tc.ticketRepo.GetAllTicketsByUserID(c,authBody.GetUser().ID)
	if err != nil {
		return jsonHelper.ApiError{
			Err:    "Error getting tickets",
			Status: 500,
		}
	}

	ticketDtos, err = tc.mapTicketList(c, &tickets)

	c.JSON(200, gin.H{"tickets":ticketDtos})

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

	fullDto,err := tc.ticketMapper.ModelToFullDto(c, &ticket)
	if err != nil {
		return jsonHelper.ApiError{
			Err:    "Failed to convert: " + err.Error(),
			Status: 500,
		}
	}

	c.JSON(200, gin.H{"ticket":fullDto})
	return nil
}
