package configs

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

/*
`ConnectMongo` connects to the MongoDB database and returns a client instance.
*/
func ConnectMongo(mongoUri string) *mongo.Client {
	client, err := mongo.NewClient(options.Client().ApplyURI(mongoUri))
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}

	// pings the database
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	defer cancel()

	fmt.Println("Connected to MongoDB")
	return client
}

// client instance
var DB *mongo.Client = ConnectMongo(os.Getenv("MONGODB_URI"))

/*
`GetCollections` returns a collection instance from the database given the collection name.
*/
func GetCollections(client *mongo.Client, collectionName string) *mongo.Collection {
	return client.Database("RealmHunter").Collection(collectionName)
}
