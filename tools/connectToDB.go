package tools

import (
	"context"
	"database/sql"
	"fmt"
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
	DB = client.Database("docker-test")

}

var RelationalDB *sql.DB

func ConnectToPostgres() {
	var err error
	username := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")
	dbname := os.Getenv("POSTGRES_DB_NAME")
	connStr := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=false",username, password, dbname )
	RelationalDB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Println("Error while connecting to utility SQL database")
		log.Fatal(err)
	}
}

