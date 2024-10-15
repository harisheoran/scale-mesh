package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	_ "github.com/joho/godotenv/autoload"
	"gitlab.com/harisheoran/scale-mesh/api-server/pkg/models"
	"gitlab.com/harisheoran/scale-mesh/api-server/pkg/models/postgresql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var dsn = os.Getenv("DBURI")

// Payload Struct for /project endpoint
type RepoUrl struct {
	GITHUB_REPO_URL string
}

type ApiConfig struct {
	address string
}

type app struct {
	errorLogger  *log.Logger
	infoLogger   *log.Logger
	projectModel *postgresql.ProjectModel
}

func main() {
	// get the Database connection pool
	dbConnectionPool, err := openDBConnectionPool(dsn)
	if err != nil {
		log.Fatal("ERROR: getting thr DB connection pool", err)
	}

	// closing the db connection pool before the main function exits.
	db, err := dbConnectionPool.DB()
	if err != nil {
		log.Fatal("ERROR: getting the DB from connection pool", err)
	}
	defer db.Close()

	projectModel := postgresql.ProjectModel{
		DBConnectionPool: dbConnectionPool,
	}

	apiConfig := ApiConfig{}
	// Create Levelled Logging
	infoLogger := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLogger := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	app := app{
		errorLogger:  errorLogger,
		infoLogger:   infoLogger,
		projectModel: &projectModel,
	}

	flag.StringVar(&apiConfig.address, "address", ":9000", "Port of the api")
	flag.Parse()

	server := &http.Server{
		Addr:     apiConfig.address,
		Handler:  app.routes(),
		ErrorLog: errorLogger,
	}

	err = server.ListenAndServe()
	if err != nil {
		app.errorLogger.Fatalf("unable to start the api at port %s", apiConfig.address)
	}

	app.infoLogger.Printf("API running on port %s", apiConfig.address)
}

func openDBConnectionPool(dsn string) (*gorm.DB, error) {
	/*
		db is here a pool of connection,
		GO manages these connection as needed, opening and closing connections to the database as needed.
		so, actual connection to the database is done lazily, as when needed for the first time.
	*/
	dbConnectionPool, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		return nil, err
	}

	// Run the automigration for Project Model
	if err := dbConnectionPool.AutoMigrate(&models.Project{}); err != nil {
		return nil, err
	}

	return dbConnectionPool, nil
}
