package main

import (
	"flag"
	"log"
	"net/http"
	"os"
)

// Payload Struct for /project endpoint
type RepoUrl struct {
	GITHUB_REPO_URL string
}

type ApiConfig struct {
	address string
}

type app struct {
	errorLogger *log.Logger
	infoLogger  *log.Logger
}

func main() {
	apiConfig := ApiConfig{}

	// Create Levelled Logging
	infoLogger := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLogger := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	app := app{
		errorLogger: errorLogger,
		infoLogger:  infoLogger,
	}

	flag.StringVar(&apiConfig.address, "address", ":9000", "Port of the api")
	flag.Parse()

	server := &http.Server{
		Addr:     apiConfig.address,
		Handler:  app.routes(),
		ErrorLog: errorLogger,
	}

	err := server.ListenAndServe()
	if err != nil {
		app.errorLogger.Fatalf("unable to start the api at port %s", apiConfig.address)
	}

	app.infoLogger.Printf("API running on port %s", apiConfig.address)
}
