package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	firebase "firebase.google.com/go"
	"github.com/99designs/gqlgen/handler"
	_ "github.com/GoogleCloudPlatform/cloudsql-proxy/proxy/dialers/postgres"
	"github.com/go-chi/chi"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/writewithwrabit/server/auth"
	graphql "github.com/writewithwrabit/server/graphql"
	"google.golang.org/api/option"
)

const defaultPort = "8080"

var db *sql.DB

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("File .env not found!")
	}

	env := os.Getenv("NODE_ENV")

	router := chi.NewRouter()

	db = DB()

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	opt := option.WithCredentialsFile(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"))
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Fatalf("error initializing app: %v\n", err)
	}

	client, err := app.Auth(context.Background())
	if err != nil {
		log.Fatalf("error getting Auth client: %v\n", err)
	}

	if env == "dev" {
		// Only allow the playground in dev
		router.Handle("/", handler.Playground("GraphQL playground", "/query"))
	} else {
		// Require a token in prod
		router.Use(auth.Middleware(client))
	}

	router.Handle("/query", handler.GraphQL(graphql.NewExecutableSchema(graphql.New(db))))

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}

// DB gets a connection to the database.
// This can panic for malformed database connection strings, invalid credentials, or non-existance database instance.
func DB() *sql.DB {
	var (
		connectionName = mustGetenv("CLOUDSQL_CONNECTION_NAME")
		user           = mustGetenv("CLOUDSQL_USER")
		dbName         = os.Getenv("CLOUDSQL_DATABASE_NAME")
		password       = os.Getenv("CLOUDSQL_PASSWORD")
		socket         = os.Getenv("CLOUDSQL_SOCKET_PREFIX")
		env            = os.Getenv("NODE_ENV")
	)

	// /cloudsql is used on App Engine.
	if socket == "" {
		socket = "/cloudsql"
	}

	dbURI := fmt.Sprintf("host=/cloudsql/%s dbname=%s user=%s password=%s", connectionName, dbName, user, password)
	dialer := "postgres"
	if env == "dev" {
		dbURI = fmt.Sprintf("host=%s dbname=%s user=%s password=%s sslmode=disable", connectionName, dbName, user, password)
		dialer = "cloudsqlpostgres"
	}

	conn, err := sql.Open(dialer, dbURI)

	if err != nil {
		panic(fmt.Sprintf("DB: %v", err))
	}

	return conn
}

func mustGetenv(k string) string {
	v := os.Getenv(k)
	if v == "" {
		log.Panicf("%s environment variable not set.", k)
	}
	return v
}
