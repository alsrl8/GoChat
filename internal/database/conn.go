package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
}

func loadPostgresConfig() *Config {
	return &Config{
		Host:     os.Getenv("LOCAL_SERVER_IP"),
		Port:     os.Getenv("POSTGRES_PORT"),
		User:     os.Getenv("POSTGRES_USER"),
		Password: os.Getenv("POSTGRES_PW"),
		Database: os.Getenv("POSTGRES_DB"),
	}
}

func loadMongoConfig() *Config {
	return &Config{
		Host:     os.Getenv("LOCAL_SERVER_IP"),
		User:     os.Getenv("MONGO_USER"),
		Password: os.Getenv("MONGO_PW"),
	}
}

func ConnectToPostgres() (*sql.DB, error) {
	cfg := loadPostgresConfig()
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Database,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("error opening postgres connection: %w", err)
	}

	if err = db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("error pinging postgres: %w", err)
	}

	log.Println("Connected to PostgreSQL")
	return db, nil
}

func ConnectToMongo() (*mongo.Client, error) {
	cfg := loadMongoConfig()
	uri := fmt.Sprintf("mongodb://%s:%s@%s", cfg.User, cfg.Password, cfg.Host)
	ctx := context.Background()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("error connecting to MongoDB: %w", err)
	}

	if err = client.Ping(ctx, nil); err != nil {
		err := client.Disconnect(context.Background())
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("error pinging MongoDB: %w", err)
	}

	log.Println("Connected to MongoDB")
	return client, nil
}
