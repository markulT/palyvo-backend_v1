package tools

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"os"
)

var DB *mongo.Database

func ConnectToDb() {

	mongoUri := os.Getenv("DB_URL")
	mongoOptions := options.Client().ApplyURI(mongoUri)
	client, err := mongo.Connect(context.TODO(), mongoOptions)
	if err != nil {
		log.Println("Error while connecting to database at /utils/connectToDb.go")
		log.Fatal(err)
	}
	DB = client.Database("palyvo-db")
	fmt.Errorf(DB.Name())
}

var RelationalDB *sql.DB
func ConnectToPostgres() {
	var err error
	username := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")
	dbname := os.Getenv("POSTGRES_DB_NAME")
	host := os.Getenv("POSTGRES_HOST")

	connStr := fmt.Sprintf("host=%s port=5432 user=%s password=%s dbname=%s sslmode=disable", host, username, password, dbname)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Printf("Error while connecting to PostgreSQL: %v", err)
		log.Fatal(err)
	}

	createTableSQL := `CREATE TABLE IF NOT EXISTS products (
		amount INTEGER,
		id UUID PRIMARY KEY,
		title TEXT,
		price INTEGER,
		currency TEXT,
		seller TEXT,
		fuel_type TEXT
	);`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		log.Printf("Error creating table: %v", err)
		log.Fatal(err)
	}

	RelationalDB = db
}



func InitSQL() {

}

type TransactionFn func(c context.Context) (interface{},error)

func WithTransaction(c context.Context, fn TransactionFn) (interface{},error) {
	sess, _ := DB.Client().StartSession()
	defer sess.EndSession(c)

	res, err := sess.WithTransaction(c, func(sc mongo.SessionContext) (interface{}, error) {
		res,err := fn(sc)
		if err != nil {
			return nil, err
		}
		return res, nil
	})
	if err != nil {
		return nil,err
	}

	return res,nil
}