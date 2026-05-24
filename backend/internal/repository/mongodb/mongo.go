package mongodb

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Client struct {
	db *mongo.Database
}

func New(uri, database string) (*Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}
	return &Client{db: client.Database(database)}, nil
}

func (c *Client) Collection(name string) *mongo.Collection {
	return c.db.Collection(name)
}

func (c *Client) Database() *mongo.Database {
	return c.db
}
