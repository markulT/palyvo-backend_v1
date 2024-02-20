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
	productTicketRepo ProductTicketRepo
}

func (cpr *consistentProductRepo) SetProductTicketRepo(repo ProductTicketRepo) {
	cpr.productTicketRepo = repo
}

func (cpr *consistentProductRepo) UpdateProductStripeID(c context.Context, pID uuid.UUID, newStripeID string) error {
	var err error
	stmt, err := tools.RelationalDB.Prepare("update products set stripe_id = $1 where id = $2")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_,err = stmt.Exec(newStripeID, pID)
	return err
}

func (cpr *consistentProductRepo) GetByFuelType(c context.Context, fuelType string) ([]models.Product, error) {
	var products []models.Product
	var err error
	stmt, err := tools.RelationalDB.Prepare("select * from products where fuel_type = $1")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	rows,err := stmt.Query(fuelType)
	if err != nil {
		return nil, err
	}
	if err = rows.Err();err!=nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var p models.Product
		err := rows.Scan(&p.Amount, &p.ID, &p.Title, &p.Price, &p.Currency, &p.Seller, &p.FuelType)
		if err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	return products, nil
}

func (cpr *consistentProductRepo) GetBySeller(c context.Context, seller string) ([]models.Product, error) {
	var products []models.Product
	var err error
	stmt, err := tools.RelationalDB.Prepare("select * from products where seller = $1")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	rows,err := stmt.Query(seller)
	if err != nil {
		return nil, err
	}
	if err = rows.Err();err!=nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var p models.Product
		err := rows.Scan(&p.Amount, &p.ID, &p.Title, &p.Price, &p.Currency, &p.Seller, &p.FuelType)
		if err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	return products, nil
}

func (cpr *consistentProductRepo) GetBySellerAndFuelType(c context.Context, seller string, fuelType string) ([]models.Product, error) {
	var products []models.Product
	var err error
	stmt, err := tools.RelationalDB.Prepare("select * from products where seller = $1 and fuel_type = $2")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	rows,err := stmt.Query(seller, fuelType)
	if err != nil {
		return nil, err
	}
	if err = rows.Err();err!=nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var p models.Product
		err := rows.Scan(&p.Amount, &p.ID, &p.Title, &p.Price, &p.Currency, &p.Seller, &p.FuelType)
		if err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	return products, nil
}

func (cpr *consistentProductRepo) GetAllProducts(c context.Context) ([]models.Product, error) {
	var products []models.Product
	var err error
	stmt, err := tools.RelationalDB.Prepare("select * from products")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	rows,err := stmt.Query()
	if err != nil {
		return nil, err
	}
	if err = rows.Err();err!=nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var p models.Product
		err := rows.Scan(&p.Amount, &p.ID, &p.Title, &p.Price, &p.Currency, &p.Seller, &p.FuelType)
		if err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	return products, nil
}

func (cpr *consistentProductRepo) UpdateProductAmount(c context.Context, pid uuid.UUID, amount int) error {
	var err error
	stmt, err := tools.RelationalDB.Prepare("update products set amount = $1 where id = $2")
	if err != nil {
		return err
	}
	err = stmt.QueryRow(amount, pid.String()).Err()
	return err
}

func (cpr *consistentProductRepo) DeleteProduct(c context.Context, pid uuid.UUID) error {
	var err error
	tx, err := tools.RelationalDB.BeginTx(c, &sql.TxOptions{Isolation: sql.LevelSerializable})

	stmt, err := tx.Prepare("delete from products where id = $1")
	if err != nil {
		tx.Rollback()
		return err
	}
	err = stmt.QueryRow(pid.String()).Err()
	if err !=nil {
		tx.Rollback()
		return err
	}

	err = cpr.productTicketRepo.DeleteManyByProductID(c, pid)
	if err != nil {
		tx.Rollback()
		return err
	}
	return err
}

func (cpr *consistentProductRepo) DecreaseProductAmount(c context.Context, pid uuid.UUID, amount int) error {
	//TODO implement me

	tx, err := tools.RelationalDB.BeginTx(c, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		tx.Rollback()
		return err
	}
	getProductStmt, err := tx.Prepare("SELECT id, amount, title from products where id = $1")
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

	updateProductStmt, err := tx.Prepare("UPDATE products SET amount = $1 WHERE id = $2")
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
	stmt, err := tools.RelationalDB.Prepare("select id, amount, title, currency,price, seller, fuel_type from products where id = $1")
	if err != nil {
		return models.Product{}, err
	}
	var stringID string
	err = stmt.QueryRow(pid.String()).Scan(&stringID, &p.Amount, &p.Title, &p.Currency, &p.Price, &p.Seller, &p.FuelType)
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
	stmt, err := tools.RelationalDB.Prepare("insert into products (amount, id, title, price, currency, seller, fuel_type) values ($1, $2, $3, $4, $5, $6, $7)")
	if err != nil {
		return err
	}
	err = stmt.QueryRow(p.Amount, p.ID.String(), p.Title, p.Price, p.Currency, p.Seller, p.FuelType).Err()
	return err
}




