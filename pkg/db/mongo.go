package db

import (
	"context"
	"log"
	"time"

	"gohighlevel/pkg/types"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDatabase struct {
	client     *mongo.Client
	collection *mongo.Collection
}

func NewMongoDatabase() *MongoDatabase {
	return &MongoDatabase{}
}

func (m *MongoDatabase) Init() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Replace with your MongoDB connection string
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	var err error
	m.client, err = mongo.Connect(ctx, clientOptions)
	if err != nil {
		return err
	}

	// Check the connection
	err = m.client.Ping(ctx, nil)
	if err != nil {
		return err
	}

	m.collection = m.client.Database("gohighlevel").Collection("samples")
	log.Println("Connected to MongoDB!")
	return nil
}

func (m *MongoDatabase) Close() {
	if m.client != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := m.client.Disconnect(ctx); err != nil {
			log.Printf("Error disconnecting from MongoDB: %v\n", err)
		}
	}
}

// InsertSample inserts a sample into the MongoDB collection.
func (m *MongoDatabase) InsertSample(sample types.Sample) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	doc := struct {
		CustomerID string    `bson:"customerId"`
		Name       string    `bson:"name"`
		Email      string    `bson:"email"`
		CreatedAt  time.Time `bson:"createdAt"`
		IngestedAt time.Time `bson:"ingestedAt"`
	}{
		CustomerID: sample.CustomerID,
		Name:       sample.Name,
		Email:      sample.Email,
		CreatedAt:  sample.CreatedAt,
		IngestedAt: time.Now(),
	}

	_, err := m.collection.InsertOne(ctx, doc)
	return err
}
