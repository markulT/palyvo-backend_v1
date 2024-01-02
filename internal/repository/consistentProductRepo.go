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

type consistentProductRepo struct {

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
	err = stmt.QueryRow(pid.String()).Scan(&p.ID, &p.Amount, &p.Title)
	if err != nil {
		return models.Product{}, err
	}
	return p ,nil
}




