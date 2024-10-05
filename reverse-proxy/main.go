package main

import (
	"log"
	"net/http"
	"strings"
)

func main() {
	// starting the server
	http.HandleFunc("/", mainHandler)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ERROR: unable to start the server", err)
	}
}

func mainHandler(w http.ResponseWriter, request *http.Request) {

	baseDomain := "https://scale-mesh-s3.s3.ap-south-1.amazonaws.com/__output/"

	hostname := request.Host
	projectID := strings.Split(hostname, ".")[0]
	redirectTo := baseDomain + projectID + "/index.html"

	log.Println("redirecting to:", redirectTo)

	http.Redirect(w, request, redirectTo, http.StatusSeeOther)

}
