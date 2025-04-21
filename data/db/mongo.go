// Package db provides database utilities for MongoDB connection and ID generation
package db

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
	"go.mongodb.org/mongo-driver/v2/mongo/writeconcern"
)

// NewMongo creates a new MongoDB connection with the given connection string and database name
// It also returns a cleanup function to close the connection
func NewMongo(dbsn, dbname string) (db *mongo.Database, cleanup func(), err error) {
	if len(dbname) == 0 || len(dbsn) == 0 {
		return nil, nil, errors.Errorf("Mongo config is empty")
	}

	var cli *mongo.Client
	cli, err = mongo.Connect(
		options.Client().ApplyURI(fmt.Sprintf("mongodb://%s", dbsn)),
		options.Client().SetWriteConcern(writeconcern.Majority()),
		options.Client().SetRetryWrites(false),
		options.Client().SetReadPreference(readpref.SecondaryPreferred()),
	)

	if err != nil {
		return nil, nil, errors.Wrapf(err, "connect to mongo failed")
	}

	cleanup = func() {
		if err = cli.Disconnect(context.Background()); err != nil {
			slog.Error("mongo disconnect failed", "error", err)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err = cli.Ping(ctx, readpref.Primary()); err != nil {
		return nil, nil, errors.Wrapf(err, "mongo ping failed")
	}

	db = cli.Database(dbname)

	return db, cleanup, nil
}

// IncrementIDDoc represents a document for storing auto-incrementing IDs
type IncrementIDDoc struct {
	Name   string `json:"name" bson:"name"`
	NextID int64  `json:"next_id" bson:"next_id"`
}

// IncrementID increments and returns the next ID for the specified collection
func IncrementID(ctx context.Context, coll *mongo.Collection, collName string) (int64, error) {
	return IncrementBatchID(ctx, coll, collName, 1)
}

// IncrementBatchID increments and returns the next batch of IDs for the specified collection
func IncrementBatchID(ctx context.Context, coll *mongo.Collection, collName string, batch int64) (int64, error) {
	if batch <= 0 {
		return 0, errors.Errorf("mongo increment batch must be greater than 0. batch=%d", batch)
	}

	result := &IncrementIDDoc{}
	if err := coll.FindOneAndUpdate(
		ctx,
		bson.M{"name": collName},
		bson.M{"$inc": bson.M{"next_id": batch}}).
		Decode(&result); err != nil {
		return 0, errors.Wrapf(err, "mongo increment batch id failed. collName=%s", collName)
	}

	return result.NextID, nil
}

// InitIncrementIDDoc initializes the increment ID document for the specified collection if it doesn't exist
func InitIncrementIDDoc(ctx context.Context, coll *mongo.Collection, incrCollName string) error {
	err := coll.FindOne(ctx, bson.M{"name": incrCollName}).Err()
	if err == nil {
		return nil
	}

	if !errors.Is(err, mongo.ErrNoDocuments) {
		return errors.Wrapf(err, "mongo find one failed. incrCollName=%s", incrCollName)
	}

	if _, err = coll.InsertOne(ctx, &IncrementIDDoc{
		Name:   incrCollName,
		NextID: 1,
	}); err != nil {
		return errors.Wrapf(err, "mongo insert one failed. incrCollName=%s", incrCollName)
	}

	slog.Info("mongo increment id doc created", "incrCollName", incrCollName)

	return nil
}
