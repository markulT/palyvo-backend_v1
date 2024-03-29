package repository

import (
	"context"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
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

	GetAll() ([]models.Ticket, error)
	GetAllTicketsByUserID(c context.Context,userID uuid.UUID) ([]models.Ticket, error)
	UpdateStatus(uuid.UUID, string) error
	UpdatePaymentID(context.Context,uuid.UUID,string) error
}

func NewTickerRepo() TicketRepo {
	return &defaultTicketRepo{}
}

type defaultTicketRepo struct {}

func (d *defaultTicketRepo) UpdatePaymentID(c context.Context,u uuid.UUID, s string) error {
	ticketCollection := tools.DB.Collection("tickets")
	_,err := ticketCollection.UpdateByID(c, u, bson.M{"$set":bson.M{"paymentId":s, "status":models.ACTIVATED}})
	if err != nil {
		return err
	}
	return nil
}

func (d *defaultTicketRepo) GetAll() ([]models.Ticket, error) {
	var tickets []models.Ticket
	ctx := context.Background()
	ticketCollection := tools.DB.Collection("tickets")
	cursor, err := ticketCollection.Find(ctx, bson.M{})
	defer cursor.Close(ctx)
	if err != nil {
		return nil, err
	}
	if cursor.Err() !=nil {
		return nil, cursor.Err()
	}
	for cursor.Next(ctx) {
		var ticket models.Ticket
		if err:=cursor.Decode(&ticket);err!=nil {
			return nil, err
		}
		tickets = append(tickets, ticket)
	}
	return tickets, nil
}

func (d *defaultTicketRepo) GetAllTicketsByUserID(c context.Context,userID uuid.UUID) ([]models.Ticket, error) {
	var tickets []models.Ticket
	ticketCollection := tools.DB.Collection("tickets")
	cursor, err := ticketCollection.Find(c, bson.M{"userId":userID})
	defer cursor.Close(c)
	if err != nil {
		return nil, err
	}
	if cursor.Err() !=nil {
		return nil, cursor.Err()
	}

	for cursor.Next(c) {
		var ticket models.Ticket
		if err:=cursor.Decode(&ticket);err!=nil {
			return nil, err
		}

		tickets = append(tickets, ticket)
	}
	return tickets, nil
}

func (d *defaultTicketRepo) UpdateStatus(u uuid.UUID, s string) error {
	ticketCollection := tools.DB.Collection("tickets")
	_,err := ticketCollection.UpdateByID(context.TODO(), u, bson.M{"$set":bson.M{"status":s}})
	return err
}

func (d *defaultTicketRepo) Create(c context.Context, ticket models.Ticket) error {
	ticketCollection := tools.DB.Collection("tickets")
	_, err := ticketCollection.InsertOne(c, ticket)
	return err
}

func (d *defaultTicketRepo) GetByID(id uuid.UUID) (models.Ticket, error) {
	var ticket models.Ticket
	ticketCollection := tools.DB.Collection("tickets")
	err := ticketCollection.FindOne(context.TODO(), bson.M{"_id":id}).Decode(&ticket)
	if err != nil {
		return models.Ticket{}, err
	}
	return ticket, nil
}

func (d *defaultTicketRepo) DeleteByID(id uuid.UUID) error {
	ticketCollection := tools.DB.Collection("tickets")
	_,err := ticketCollection.DeleteOne(context.TODO(), bson.M{"_id":id})
	return err
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