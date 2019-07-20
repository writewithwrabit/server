package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/handler"
	_ "github.com/GoogleCloudPlatform/cloudsql-proxy/proxy/dialers/postgres"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	graphql "github.com/writewithwrabit/server/graphql"
)

const defaultPort = "8080"

var db *sql.DB

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("File .env not found!")
	}

	db = DB()

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	http.Handle("/", handler.Playground("GraphQL playground", "/query"))
	http.Handle("/query", handler.GraphQL(graphql.NewExecutableSchema(graphql.Config{Resolvers: &graphql.Resolver{}})))

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
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
		env            = os.Getenv("ENV")
	)

	// /cloudsql is used on App Engine.
	if socket == "" {
		socket = "/cloudsql"
	}

	dbURI := fmt.Sprintf("host=%s dbname=%s user=%s password=%s sslmode=disable", connectionName, dbName, user, password)
	dialer := "postgres"
	if env == "dev" {
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
