package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	_ "github.com/joho/godotenv/autoload"
	"github.com/redis/go-redis/v9"
)

var (
	redisClient = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("ADDRESS"),
		Password: os.Getenv("PASSWORD"),
		DB:       0,
	})

	ctx      = context.Background()
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	projectID = "zqd8k"
)

func main() {

	// Load env
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	_, err = redisClient.Ping(ctx).Result()
	if err != nil {
		log.Println("ERROR: Unable to connect with Redis.", err)
	}

	http.HandleFunc("/", websockerHandler)

	log.Println("Startin the server")
	err = http.ListenAndServe(":9001", nil)
	if err != nil {
		log.Fatal("Unable to start the server on port 9001")
	}

}

func websockerHandler(w http.ResponseWriter, request *http.Request) {
	// upgrade the connection if upgrader header is in the request
	wsConnection, err := upgrader.Upgrade(w, request, nil)
	if err != nil {
		log.Println("ERROR: upgrading the connection to web socket", err)
		return
	}

	defer wsConnection.Close()

	channel := fmt.Sprintf("logs:%s", projectID)
	log.Println("Listening on channel: ", channel)
	subscribe := redisClient.Subscribe(ctx, channel)

	defer subscribe.Close()

	// keep the connection alive
	for {
		message, err := subscribe.ReceiveMessage(ctx)
		if err != nil {
			log.Println("ERROR: unable to recieve message from the channel")
		}

		log.Println("LOGS: ", message.Channel, message.Payload)

	}
}
