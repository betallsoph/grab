package db

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func NewMongo(uri string) *mongo.Client {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatalf("[mongo] failed to connect: %v", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		log.Fatalf("[mongo] ping failed: %v", err)
	}
	log.Println("[mongo] connected")
	return client
}
