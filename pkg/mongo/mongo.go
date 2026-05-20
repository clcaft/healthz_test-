package mongo

import (
	"context"
	"fmt"
	"net/url"
	"path"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"gl.eda1.ru/go/go-service-template/pkg/logger"
)

// Server is exported.
type Server struct {
	Client  *mongo.Client
	DB      *mongo.Database
	context context.Context

	error chan error
	stop  chan struct{}

	timeout time.Duration
	logger  logger.Interface

	Cancel func()
}

// New is exported.
func New(ctx context.Context, uri string, timeout time.Duration, l logger.Interface) (*Server, error) {
	var err error

	mongoServer := &Server{
		context: ctx,
		error:   make(chan error),
		stop:    make(chan struct{}),
		timeout: timeout,
		logger:  l,
	}

	clientOpts := options.Client().ApplyURI(uri).SetTimeout(timeout)

	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return mongoServer, fmt.Errorf("mongo - NewMongo - Database connection error: %w", err)
	}

	mongoServer.Client = client

	// ping
	err = mongoServer.Client.Ping(ctx, readpref.Primary())
	if err != nil {
		return mongoServer, fmt.Errorf("mongo - NewMongo - Ping error: %w", err)
	}

	u, err := url.Parse(uri)
	if err != nil {
		return mongoServer, fmt.Errorf("mongo - NewMongo - Database error: %w", err)
	}

	database := path.Base(u.Path)
	if database == "/" || database == "." {
		return mongoServer, fmt.Errorf("mongo - NewMongo - Database error: database not set")
	}

	mongoServer.DB = mongoServer.Client.Database(database)

	return mongoServer, nil
}

func (c *Server) Shutdown() error {
	select {
	case <-c.error:
		return nil
	default:
	}

	close(c.stop)
	time.Sleep(c.timeout)

	err := c.Client.Disconnect(context.Background())
	if err != nil {
		return fmt.Errorf("mongoServer - Server - Shutdown - s.Client.Disconnect: %w", err)
	}

	return nil
}
