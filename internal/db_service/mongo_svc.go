package db_service

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DbService[DocType interface{}] interface {
	CreateDocument(ctx context.Context, id string, document *DocType) error
	FindDocument(ctx context.Context, id string) (*DocType, error)
	UpdateDocument(ctx context.Context, id string, document *DocType) error
	DeleteDocument(ctx context.Context, id string) error
	Disconnect(ctx context.Context) error
}

var ErrNotFound = fmt.Errorf("document not found")
var ErrConflict = fmt.Errorf("conflict: document already exists")

type MongoServiceConfig struct {
	ServerHost string
	ServerPort int
	UserName   string
	Password   string
	DbName     string
	Collection string
	Timeout    time.Duration
}

type mongoSvc[DocType interface{}] struct {
	MongoServiceConfig
	client     atomic.Pointer[mongo.Client]
	clientLock sync.Mutex
}

func NewMongoService[DocType interface{}](config MongoServiceConfig) DbService[DocType] {
	enviro := func(name string, defaultValue string) string {
		if value, ok := os.LookupEnv(name); ok {
			return value
		}
		return defaultValue
	}

	svc := &mongoSvc[DocType]{}
	svc.MongoServiceConfig = config

	if svc.ServerHost == "" {
		svc.ServerHost = enviro("AMBULANCE_API_MONGODB_HOST", "localhost")
	}

	if svc.ServerPort == 0 {
		port := enviro("AMBULANCE_API_MONGODB_PORT", "27017")
		if p, err := strconv.Atoi(port); err == nil {
			svc.ServerPort = p
		} else {
			log.Printf("Invalid port value: %v", port)
			svc.ServerPort = 27017
		}
	}

	if svc.UserName == "" {
		svc.UserName = enviro("AMBULANCE_API_MONGODB_USERNAME", "")
	}

	if svc.Password == "" {
		svc.Password = enviro("AMBULANCE_API_MONGODB_PASSWORD", "")
	}

	if svc.DbName == "" {
		svc.DbName = enviro("AMBULANCE_API_MONGODB_DATABASE", "scs-ambulance-wl")
	}

	if svc.Collection == "" {
		svc.Collection = enviro("AMBULANCE_API_MONGODB_COLLECTION", "ambulance")
	}

	if svc.Timeout == 0 {
		seconds := enviro("AMBULANCE_API_MONGODB_TIMEOUT_SECONDS", "10")
		if s, err := strconv.Atoi(seconds); err == nil {
			svc.Timeout = time.Duration(s) * time.Second
		} else {
			log.Printf("Invalid timeout value: %v", seconds)
			svc.Timeout = 10 * time.Second
		}
	}

	log.Printf(
		"MongoDB config: //%v@%v:%v/%v/%v",
		svc.UserName,
		svc.ServerHost,
		svc.ServerPort,
		svc.DbName,
		svc.Collection,
	)
	return svc
}

func (this *mongoSvc[DocType]) connect(ctx context.Context) (*mongo.Client, error) {
	client := this.client.Load()
	if client != nil {
		return client, nil
	}

	this.clientLock.Lock()
	defer this.clientLock.Unlock()
	client = this.client.Load()
	if client != nil {
		return client, nil
	}

	ctx, contextCancel := context.WithTimeout(ctx, this.Timeout)
	defer contextCancel()

	uri := fmt.Sprintf("mongodb://%v:%v", this.ServerHost, this.ServerPort)
	log.Printf("Using URI: " + uri)

	if len(this.UserName) != 0 {
		uri = fmt.Sprintf("mongodb://%v:%v@%v:%v", this.UserName, this.Password, this.ServerHost, this.ServerPort)
	}

	if c, err := mongo.Connect(ctx, options.Client().ApplyURI(uri).SetConnectTimeout(10*time.Second)); err != nil {
		return nil, err
	} else {
		this.client.Store(c)
		return c, nil
	}
}

func (this *mongoSvc[DocType]) Disconnect(ctx context.Context) error {
	client := this.client.Load()

	if client != nil {
		this.clientLock.Lock()
		defer this.clientLock.Unlock()

		client = this.client.Load()
		defer this.client.Store(nil)
		if client != nil {
			if err := client.Disconnect(ctx); err != nil {
				return err
			}
		}
	}
	return nil
}

func (this *mongoSvc[DocType]) CreateDocument(ctx context.Context, id string, document *DocType) error {
	ctx, contextCancel := context.WithTimeout(ctx, this.Timeout)
	defer contextCancel()
	client, err := this.connect(ctx)
	if err != nil {
		return err
	}
	db := client.Database(this.DbName)
	collection := db.Collection(this.Collection)
	result := collection.FindOne(ctx, bson.D{{Key: "id", Value: id}})
	switch result.Err() {
	case nil:
		return ErrConflict
	case mongo.ErrNoDocuments:
		// expected
	default:
		return result.Err()
	}

	_, err = collection.InsertOne(ctx, document)
	return err
}

func (this *mongoSvc[DocType]) FindDocument(ctx context.Context, id string) (*DocType, error) {
	ctx, contextCancel := context.WithTimeout(ctx, this.Timeout)
	defer contextCancel()
	client, err := this.connect(ctx)
	if err != nil {
		return nil, err
	}
	db := client.Database(this.DbName)
	collection := db.Collection(this.Collection)
	result := collection.FindOne(ctx, bson.D{{Key: "id", Value: id}})
	switch result.Err() {
	case nil:
	case mongo.ErrNoDocuments:
		return nil, ErrNotFound
	default:
		return nil, result.Err()
	}
	var document *DocType
	if err := result.Decode(&document); err != nil {
		return nil, err
	}
	return document, nil
}

func (this *mongoSvc[DocType]) UpdateDocument(ctx context.Context, id string, document *DocType) error {
	ctx, contextCancel := context.WithTimeout(ctx, this.Timeout)
	defer contextCancel()
	client, err := this.connect(ctx)
	if err != nil {
		return err
	}
	db := client.Database(this.DbName)
	collection := db.Collection(this.Collection)
	result := collection.FindOne(ctx, bson.D{{Key: "id", Value: id}})
	switch result.Err() {
	case nil:
	case mongo.ErrNoDocuments:
		return ErrNotFound
	default:
		return result.Err()
	}
	_, err = collection.ReplaceOne(ctx, bson.D{{Key: "id", Value: id}}, document)
	return err
}

func (this *mongoSvc[DocType]) DeleteDocument(ctx context.Context, id string) error {
	ctx, contextCancel := context.WithTimeout(ctx, this.Timeout)
	defer contextCancel()
	client, err := this.connect(ctx)
	if err != nil {
		return err
	}
	db := client.Database(this.DbName)
	collection := db.Collection(this.Collection)
	result := collection.FindOne(ctx, bson.D{{Key: "id", Value: id}})
	switch result.Err() {
	case nil:
	case mongo.ErrNoDocuments:
		return ErrNotFound
	default:
		return result.Err()
	}
	_, err = collection.DeleteOne(ctx, bson.D{{Key: "id", Value: id}})
	return err
}
