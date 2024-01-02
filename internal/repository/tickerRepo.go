package repository

import (
	"context"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/mongo"
	"palyvoua/internal/models"
	"palyvoua/tools"
)

type CreateTicketTransactionFn func(c context.Context) error

type TicketRepo interface {
	Create(context.Context, models.Ticket) error
	GetByID(id uuid.UUID) (models.Ticket, error)
	DeleteByID(id uuid.UUID) error
	WithTransaction(c context.Context, fn CreateTicketTransactionFn) error
}

func NewTickerRepo() TicketRepo {
	return &defaultTicketRepo{}
}

type defaultTicketRepo struct {}

func (d *defaultTicketRepo) Create(c context.Context, ticket models.Ticket) error {
	//TODO implement me
	return nil
}

func (d *defaultTicketRepo) GetByID(id uuid.UUID) (models.Ticket, error) {
	//TODO implement me
	return models.Ticket{}, nil
}

func (d *defaultTicketRepo) DeleteByID(id uuid.UUID) error {
	return nil
}

func (d *defaultTicketRepo) WithTransaction(c context.Context, fn CreateTicketTransactionFn) error {

	sess, _ := tools.DB.Client().StartSession()
	defer sess.EndSession(c)

	_, err := sess.WithTransaction(c, func(sc mongo.SessionContext) (interface{}, error) {
		err := fn(sc)
		if err != nil {
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		return err
	}

	return nil
}