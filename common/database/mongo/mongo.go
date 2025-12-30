package mongo

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// MongoDBAdapter struct.
// This struct is used to represent a MongoDB adapter.
//
// Attributes:
//   - client (*mongo.Client): The MongoDB client.
//   - database (*mongo.Database): The MongoDB database.
type MongoDBAdapter struct {
	client   *mongo.Client
	database *mongo.Database
}

// NewMongoDBAdapter function.
// This function is used to create a new MongoDB adapter.
//
// Parameters:
//   - uri (string): The MongoDB URI.
//   - dbName (string): The MongoDB database name.
//
// Returns:
//   - *MongoDBAdapter: The MongoDB adapter.
//   - error: The error.
func NewMongoDBAdapter(uri, dbName string) (*MongoDBAdapter, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, err
	}

	log.Println("Connected to MongoDB!")

	return &MongoDBAdapter{
		client:   client,
		database: client.Database(dbName),
	}, nil
}

// Collection function.
// This function is used to get a collection from the MongoDB database.
//
// Parameters:
//   - name (string): The collection name.
//
// Returns:
//   - *mongo.Collection: The MongoDB collection.
func (adapter *MongoDBAdapter) Collection(name string) *mongo.Collection {
	return adapter.database.Collection(name)
}

// Disconnect function.
// This function is used to disconnect the MongoDB client.
//
// Parameters:
//   - ctx (context.Context): The context.
//
// Returns:
//   - error: The error.
func (adapter *MongoDBAdapter) Disconnect(ctx context.Context) error {
	return adapter.client.Disconnect(ctx)
}
