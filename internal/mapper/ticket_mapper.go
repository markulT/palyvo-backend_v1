package mapper

import (
	"context"
	"palyvoua/internal/dto"
	"palyvoua/internal/models"
	"palyvoua/internal/repository"
)

type TicketMapper interface {
	ProductTicketToDto(model *models.ProductTicket) (*dto.ProductTicketDto, error)
	ModelToFullDto(c context.Context,model *models.Ticket) (*dto.FullTicketDto, error)
}

type defaultTicketMapper struct {
	ptr repository.ProductTicketRepo
	tr repository.TicketRepo
}

func (d *defaultTicketMapper) ProductTicketToDto(model *models.ProductTicket) (*dto.ProductTicketDto, error) {
	amount := model.Amount
	price := model.Price
	ptDto:= dto.ProductTicketDto{
		ProductID: stringPtr(model.ProductID.String()),
		Amount:    &amount,
		Title:     stringPtr(model.Title),
		Price:     &price,
		Currency:  stringPtr(model.Currency),
		StripeID:  stringPtr(model.StripeID),
		Seller:    stringPtr(model.Seller),
		FuelType:  stringPtr(model.FuelType),
	}
	return &ptDto,nil
}

func (d *defaultTicketMapper) TicketToDto(model *models.Ticket) (*dto.TicketDto, error) {
	createdAt := model.CreatedAt
	expiresAt := model.ExpiresAt
	amount := model.Amount
	ticketDto := dto.TicketDto{
		CreatedAt:       &createdAt,
		ExpiresAt:       &expiresAt,
		ID:              stringPtr(model.ID.String()),
		UserId:          stringPtr(model.UserId.String()),
		Status:          stringPtr(model.Status),
		Amount:          &amount,
		PaymentID:       stringPtr(model.PaymentID),
		ProductTicketId: stringPtr(model.ProductTicketID.String()),
	}
	return &ticketDto,nil
}

func (d *defaultTicketMapper) ModelToFullDto(c context.Context,model *models.Ticket) (*dto.FullTicketDto, error) {
	var fullTicketDto dto.FullTicketDto
	var err error

	productTicket, err := d.ptr.GetByID(c, model.ProductTicketID)
	if err != nil {
		return nil, err
	}

	ptDto, err := d.ProductTicketToDto(&productTicket)
	if err != nil {
		return nil, err
	}

	fullTicketDto.ProductTicket = *ptDto
	ticketDto, err := d.TicketToDto(model)

	fullTicketDto.TicketDto = *ticketDto

	return &fullTicketDto,nil
}

type TicketMapperOptions struct {
	ProductTicketRepo repository.ProductTicketRepo
	TicketRepo repository.TicketRepo
}

func NewTicketMapper(options TicketMapperOptions) TicketMapper {
	mapper := defaultTicketMapper{}
	mapper.ptr = options.ProductTicketRepo
	return &mapper
}
