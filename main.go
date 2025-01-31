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
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	_ "github.com/sqreen/go-agent/agent"
	"github.com/sqreen/go-agent/sdk/middleware/sqhttp"
	"github.com/writewithwrabit/server/auth"
	"github.com/writewithwrabit/server/graph/generated"
	"github.com/writewithwrabit/server/resolvers"
	"google.golang.org/api/option"
)

const defaultPort = "8080"

var db *sql.DB

func main() {
	env := os.Getenv("NODE_ENV")
	if "" == env {
		env = "dev"
	}

	if err := godotenv.Load("." + env + ".env"); err != nil {
		log.Println("File .env not found!")
	}

	router := chi.NewRouter()

	// Basic CORS
	// for more ideas, see: https://developer.github.com/v3/#cross-origin-resource-sharing
	cors := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	})
	router.Use(cors.Handler)

	db = DB()

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	// Setup Google token verification
	opt := option.WithCredentialsFile(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"))
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Fatalf("error initializing app: %v\n", err)
	}

	client, err := app.Auth(context.Background())
	if err != nil {
		log.Fatalf("error getting Auth client: %v\n", err)
	}

	router.Use(auth.Middleware(client))

	router.Handle("/query", handler.GraphQL(
		generated.NewExecutableSchema(resolvers.New(db))),
	)

	if env == "dev" {
		// Only allow the playground in dev
		router.Handle("/", handler.Playground("GraphQL playground", "/query"))
		log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)

		log.Fatal(http.ListenAndServe(":"+port, router))
	} else {
		log.Fatal(http.ListenAndServe(":"+port, sqhttp.Middleware(router)))
	}
}

// DB gets a connection to the database.
// This can panic for malformed database connection strings, invalid credentials, or non-existance database instance.
func DB() *sql.DB {
	var (
		connectionName = mustGetenv("CLOUDSQL_CONNECTION_NAME")
		user           = mustGetenv("CLOUDSQL_USER")
		dbName         = os.Getenv("CLOUDSQL_DATABASE_NAME")
		password       = os.Getenv("CLOUDSQL_PASSWORD")
		env            = os.Getenv("NODE_ENV")
	)

	dbURI := fmt.Sprintf("host=/cloudsql/%s dbname=%s user=%s password=%s", connectionName, dbName, user, password)
	dialer := "postgres"
	if env == "dev" {
		dbURI = fmt.Sprintf("host=%s dbname=%s user=%s password=%s sslmode=disable", connectionName, dbName, user, password)
		// dialer = "cloudsqlpostgres"
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
