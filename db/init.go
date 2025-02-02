package db

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"imdb/logger"
	"imdb/model"
	"log"
	"os"
	"time"
)

var (
	MongoUri = "mongodb://127.0.0.1:27017"
	Client   *mongo.Client
	err      error
	Opts     *options.ReplaceOptions
)

func init() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	Client, err = mongo.Connect(ctx, options.Client().ApplyURI(MongoUri))
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	err = Client.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	Opts = options.Replace().SetUpsert(true)
}

func DisconnectDB() {
	if err = Client.Disconnect(context.TODO()); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}

func ListDatabases() []string {
	databases, err := Client.ListDatabaseNames(context.TODO(), bson.M{})
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	return databases
}

func ReplaceOne(data model.Movie) {
	if _, err = Client.Database("imdb").Collection("movie").ReplaceOne(context.TODO(), bson.M{"id": data.ID}, data, Opts); err != nil {
		logger.WriteLog(fmt.Sprintln(time.Now().Format(time.RFC1123), "[ERR]", err))
	} else {
		fmt.Println(data.ID, "saved")
	}
}
