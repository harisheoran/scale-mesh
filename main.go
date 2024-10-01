package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {

  http.HandleFunc("/", rootHandler)

  err := http.ListenAndServe(":8080", nil)
  if err != nil {
    log.Fatal("Error Starting the server")
  }
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
  fmt.Fprintf(w, "Welcome to the Scale Mesh")
}

