package mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var ErrMissingURI = &ConfigError{"TUNNELIER_MONGO_URL not set"}

type ConfigError struct {
	Msg string
}

type VPNConfig struct {
	Name   string `bson:"name"`
	Type   string `bson:"type"`
	Config string `bson:"config"`
}

func (e *ConfigError) Error() string {
	return e.Msg
}

type Store struct {
	Client     *mongo.Client
	DBName     string
	Collection string
}

func NewStore(ctx context.Context, mongoUrl, dbName, collection string) (*Store, error) {
	if mongoUrl == "" {
		return nil, ErrMissingURI
	}

	opts := options.Client().ApplyURI(mongoUrl)
	client, err := mongo.Connect(opts)

	if err != nil {
		return nil, err
	}

	return &Store{
		Client:     client,
		DBName:     dbName,
		Collection: collection,
	}, nil
}

func (s *Store) FetchWireguardSample(ctx context.Context, count int) ([]VPNConfig, error) {
	col := s.Client.Database(s.DBName).Collection(s.Collection)

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.D{{Key: "type", Value: "wireguard"}}}},
		{{Key: "$sample", Value: bson.D{{Key: "size", Value: count}}}},
	}

	cursor, err := col.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []VPNConfig
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}

func (s *Store) StoreWireguardConfig(ctx context.Context, cfg VPNConfig) error {
	col := s.Client.Database(s.DBName).Collection(s.Collection)

	_, err := col.InsertOne(ctx, cfg)
	return err
}
