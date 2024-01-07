package repository

import (
	"context"
	"database/sql"
	"github.com/google/uuid"
	"palyvoua/internal/models"
	"palyvoua/tools"
)

func NewConsistentProductRepo() ProductRepo {
	return &consistentProductRepo{}
}

type consistentProductRepo struct {}

func (cpr *consistentProductRepo) UpdateProductAmount(c context.Context, pid uuid.UUID, amount int) error {
	var err error
	stmt, err := tools.RelationalDB.Prepare("update product set amount = ? where id = ?")
	if err != nil {
		return err
	}
	err = stmt.QueryRow(amount, pid.String()).Err()
	return err
}

func (cpr *consistentProductRepo) DeleteProduct(c context.Context, pid uuid.UUID) error {
	var err error
	stmt, err := tools.RelationalDB.Prepare("delete from product where id = ?")
	if err != nil {
		return err
	}
	err = stmt.QueryRow(pid.String()).Err()
	return err
}

func (cpr *consistentProductRepo) DecreaseProductAmount(c context.Context, pid uuid.UUID, amount int) error {
	//TODO implement me

	tx, err := tools.RelationalDB.BeginTx(nil, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		tx.Rollback()
		return err
	}
	getProductStmt, err := tx.Prepare("SELECT id, amount, title from products where id = ?")
	if err != nil {
		tx.Rollback()
		return err
	}
	p := models.Product{}
	err = getProductStmt.QueryRow(pid.String()).Scan(&p.ID, &p.Amount, &p.Title)
	if err != nil {
		tx.Rollback()
		return err
	}

	newAmount := p.Amount - amount

	updateProductStmt, err := tx.Prepare("UPDATE products SET amount = ? WHERE id = ?")
	if err != nil {
		tx.Rollback()
		return err
	}
	_, err = updateProductStmt.Exec(newAmount, p.ID)

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return err
	}
	return nil
}

func (cpr *consistentProductRepo) GetProduct(c context.Context, pid uuid.UUID) (models.Product, error) {
	var err error
	p := models.Product{}
	stmt, err := tools.RelationalDB.Prepare("select id, amount, title from products where id = ?")
	if err != nil {
		return models.Product{}, err
	}
	var stringID string
	err = stmt.QueryRow(pid.String()).Scan(&stringID, &p.Amount, &p.Title)
	if err != nil {
		return models.Product{}, err
	}
	p.ID, err = uuid.Parse(stringID)
	if err != nil {
		return models.Product{}, err
	}
	return p ,nil
}

func (cpr *consistentProductRepo) SaveProduct(c context.Context, p *models.Product) error {
	var err error
	stmt, err := tools.RelationalDB.Prepare("insert into products (amount, id, title, price, currency) values (?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}
	err = stmt.QueryRow(p.Amount, p.ID.String(), p.Title, p.Price, p.Currency).Err()
	return err
}




