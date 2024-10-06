package main

import (
	"bytes"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
)

type DeployResponse struct {
	Status     string
	WebsiteUrl string
}

var apiUrl = "http://127.0.0.1:9000/project"

func main() {
	http.HandleFunc("/", homePage)

	err := http.ListenAndServe(":4000", nil)
	if err != nil {
		log.Fatal("ERROR: starting the server ", err)
	}

}

func homePage(w http.ResponseWriter, request *http.Request) {
	var deployResponse DeployResponse
	uiTemplate := "ui/index.gohtml"
	template, err := template.ParseFiles(uiTemplate)
	if err != nil {
		log.Fatal("ERROR: parsing the ui template files", err)
	}

	if request.Method == http.MethodPost {
		// send the json data to backend
		payloadURL := request.FormValue("url")

		payloadToSend := map[string]string{
			"GITHUB_REPO_URL": payloadURL,
		}

		payloadToSendJSON, err := json.Marshal(payloadToSend)
		if err != nil {
			log.Fatal("ERROR: unable to convert payload to JSON")
		}

		response, err := http.Post(apiUrl, "application/json", bytes.NewBuffer(payloadToSendJSON))
		if err != nil {
			log.Fatal("ERROR: unable to send payload to api server")
		}
		defer response.Body.Close()

		// show response from backend to users
		err = json.NewDecoder(response.Body).Decode(&deployResponse)
		if err != nil {
			log.Println("ERROR: parsing the json response", err)
		}
	}

	template.Execute(w, deployResponse)
}
