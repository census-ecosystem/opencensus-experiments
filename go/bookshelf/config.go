// Copyright 2015 Google Inc. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package bookshelf

import (
	"log"
	"os"

	"cloud.google.com/go/datastore"
	"cloud.google.com/go/pubsub"
	"cloud.google.com/go/storage"

	"gopkg.in/mgo.v2"

	"github.com/gorilla/sessions"

	"contrib.go.opencensus.io/exporter/stackdriver"
	"go.opencensus.io/plugin/ocgrpc"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var (
	DB          BookDatabase
	OAuthConfig *oauth2.Config

	StorageBucket     *storage.BucketHandle
	StorageBucketName string

	SessionStore sessions.Store

	PubsubClient *pubsub.Client

	// Force import of mgo library.
	_ mgo.Session
)

const PubsubTopicID = "fill-book-details"
const projectID = "bookshelf-195421"

func init() {
	var err error

	exporter, err := stackdriver.NewExporter(stackdriver.Options{ProjectID: projectID})
	if err != nil {
		log.Fatal(err)
	}
	trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})
	trace.RegisterExporter(exporter)

	view.RegisterExporter(exporter)

	// register to views
	view.Register(ochttp.DefaultServerViews...)
	view.Register(ochttp.DefaultClientViews...)
	view.Register(ocgrpc.DefaultServerViews...)
	view.Register(ocgrpc.DefaultClientViews...)

	log.Printf("installed opencensus trace exporter")

	span := trace.NewSpan("test-span-"+os.Args[0], nil, trace.StartOptions{})
	span.End()

	// To use the in-memory test database, uncomment the next line.
	//DB = newMemoryDB()

	// [START cloudsql]
	// To use Cloud SQL, uncomment the following lines, and update the username,
	// password and instance connection string. When running locally,
	// localhost:3306 is used, and the instance name is ignored.
	// DB, err = configureCloudSQL(cloudSQLConfig{
	// 	Username: "root",
	// 	Password: "",
	// 	// The connection name of the Cloud SQL v2 instance, i.e.,
	// 	// "project:region:instance-id"
	// 	// Cloud SQL v1 instances are not supported.
	// 	Instance: "",
	// })
	// [END cloudsql]

	// [START mongo]
	// To use Mongo, uncomment the next lines and update the address string and
	// optionally, the credentials.
	//
	// var cred *mgo.Credential
	// DB, err = newMongoDB("localhost", cred)
	// [END mongo]

	// [START datastore]
	// To use Cloud Datastore, uncomment the following lines and update the
	// project ID.
	// More options can be set, see the google package docs for details:
	// http://godoc.org/golang.org/x/oauth2/google
	//
	DB, err = configureDatastoreDB(projectID)
	// [END datastore]

	if err != nil {
		log.Fatal(err)
	}

	// [START storage]
	// To configure Cloud Storage, uncomment the following lines and update the
	// bucket name.
	//
	StorageBucketName = projectID
	StorageBucket, err = configureStorage(StorageBucketName)
	// [END storage]

	if err != nil {
		log.Fatal(err)
	}

	// [START auth]
	// To enable user sign-in, uncomment the following lines and update the
	// Client ID and Client Secret.
	// You will also need to update OAUTH2_CALLBACK in app.yaml when pushing to
	// production.
	//
	// OAuthConfig = configureOAuthClient("clientid", "clientsecret")
	// [END auth]

	// [START sessions]
	// Configure storage method for session-wide information.
	// Update "something-very-secret" with a hard to guess string or byte sequence.
	cookieStore := sessions.NewCookieStore([]byte("something-very-secret"))
	cookieStore.Options = &sessions.Options{
		HttpOnly: true,
	}
	SessionStore = cookieStore
	// [END sessions]

	// [START pubsub]
	// To configure Pub/Sub, uncomment the following lines and update the project ID.
	//
	PubsubClient, err = configurePubsub(projectID)
	// [END pubsub]

	if err != nil {
		log.Fatal(err)
	}
}

func configureDatastoreDB(projectID string) (BookDatabase, error) {
	ctx := context.Background()
	client, err := datastore.NewClient(ctx, projectID)
	if err != nil {
		return nil, err
	}
	return newDatastoreDB(client)
}

func configureStorage(bucketID string) (*storage.BucketHandle, error) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	return client.Bucket(bucketID), nil
}

func configurePubsub(projectID string) (*pubsub.Client, error) {
	//if _, ok := DB.(*memoryDB); ok {
	//	return nil, errors.New("Pub/Sub worker doesn't work with the in-memory DB " +
	//		"(worker does not share its memory as the main app). Configure another " +
	//		"database in bookshelf/config.go first (e.g. MySQL, Cloud Datastore, etc)")
	//}

	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return nil, err
	}

	// Create the topic if it doesn't exist.
	if exists, err := client.Topic(PubsubTopicID).Exists(ctx); err != nil {
		return nil, err
	} else if !exists {
		if _, err := client.CreateTopic(ctx, PubsubTopicID); err != nil {
			return nil, err
		}
	}
	return client, nil
}

func configureOAuthClient(clientID, clientSecret string) *oauth2.Config {
	redirectURL := os.Getenv("OAUTH2_CALLBACK")
	if redirectURL == "" {
		redirectURL = "http://localhost:8080/oauth2callback"
	}
	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       []string{"email", "profile"},
		Endpoint:     google.Endpoint,
	}
}

type cloudSQLConfig struct {
	Username, Password, Instance string
}

//func configureCloudSQL(config cloudSQLConfig) (BookDatabase, error) {
//	if os.Getenv("GAE_INSTANCE") != "" {
//		// Running in production.
//		return newMySQLDB(MySQLConfig{
//			Username:   config.Username,
//			Password:   config.Password,
//			UnixSocket: "/cloudsql/" + config.Instance,
//		})
//	}
//
//	// Running locally.
//	return newMySQLDB(MySQLConfig{
//		Username: config.Username,
//		Password: config.Password,
//		Host:     "localhost",
//		Port:     3306,
//	})
//}
