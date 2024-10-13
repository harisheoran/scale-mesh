package main

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"mime"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/joho/godotenv"
	_ "github.com/joho/godotenv/autoload"
	"github.com/redis/go-redis/v9"
)

var redisClient = redis.NewClient(&redis.Options{
	Addr:     os.Getenv("ADDRESS"),
	Password: os.Getenv("PASSWORD"),
	DB:       0,
})

var ctx = context.Background()

var projectID = os.Getenv("projectID")

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	_, err = redisClient.Ping(ctx).Result()
	if err != nil {
		log.Println("ERROR: Unable to connect with Redis.", err)
	}
	log.Println("Uploading logs to the channel", fmt.Sprintf("logs:%s", projectID))

	// Get the project ID
	fmt.Println("Project ID", projectID)
	publishLogs(fmt.Sprintf("Project ID %s", projectID))

	// Authenticate with AWS SDK
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("ap-south-1"))
	if err != nil {
		publishLogs(createErrorLogs("unable to auth with AWS", err))
		log.Fatal("ERROR: unable to auth with AWS", err)
	}

	// Create a S3 client
	client := s3.NewFromConfig(cfg)

	/*
	   1. Build the code
	*/

	// Go to the clonned repo and build the code
	path := "/app/output"
	err = os.Chdir(path)
	if err != nil {
		publishLogs(createErrorLogs("unable to find the cloned repository, repo is not cloned", err))
		log.Fatal("ERROR: unable to find cloned repo", err)
	}

	// build the application
	log.Println("Building the application ...")
	publishLogs(createInfoLogs("Building the application ..."))
	cmd := exec.Command("/bin/sh", "-c", "npm install && npm run build")

	_, err = cmd.CombinedOutput()
	if err != nil {
		publishLogs(createErrorLogs("unable to build the application.", err))
		log.Fatal("ERROR: unable to build the application", err)
	}
	log.Println("Build completed...")
	publishLogs(createInfoLogs("Build completed."))

	/*
	   2. Upload the build artifacts to S3 buckets.
	*/

	log.Println("Uploading build artifacts to S3 bucket...")
	publishLogs(createInfoLogs("Uploading the build artifacts to the S3 bucket..."))
	// change the directory to the build output directory
	buildOutputPath := "/app/output/dist/"
	err = os.Chdir(buildOutputPath)
	if err != nil {
		publishLogs(createErrorLogs("Build directory not found /app/output/dist", err))
		log.Fatal("ERROR: Build directory not found", err)
	}

	// go through file recursively
	err = filepath.Walk(".",
		func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// check for directory: we dont want to upload the directory, only the files within
			if !info.IsDir() {
				err = uploadArtifactToS3(client, path, projectID)
				if err != nil {
					publishLogs(createErrorLogs("unable to upload the build artifacts", err))
					log.Fatal("ERROR: uploading the file to S3", err)
				}
			}
			return nil
		},
	)
	if err != nil {
		publishLogs(createErrorLogs("unable to upload the build artifacts", err))
		fmt.Println(err)
	}

	publishLogs(createInfoLogs("Build artifacts are uploaded successfully."))
}

// Upload artifacts to s3 buckets
func uploadArtifactToS3(s3Client *s3.Client, filename string, projectID string) error {
	var bucketName = "scale-mesh-s3"
	var objectKey = "__output/" + projectID + "/" + filename

	file, err := os.Open(filename)
	if err != nil {
		return err
	}

	defer file.Close()

	ext := filepath.Ext(filename)
	contentType := mime.TypeByExtension(ext)
	log.Printf("Uploading %s with content-type: %s", filename, contentType)

	_, err = s3Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String(bucketName),
		Body:        file,
		Key:         aws.String(objectKey),
		ContentType: aws.String(contentType),
	})

	if err != nil {
		return err
	}

	return nil
}

func publishLogs(log string) error {
	channel := fmt.Sprintf("logs:%s", projectID)

	err := redisClient.Publish(ctx, channel, log).Err()
	if err != nil {
		return err
	}

	return nil
}

func createErrorLogs(log string, err error) string {
	return fmt.Sprintf("ERROR: %s, %s", log, err)
}

func createInfoLogs(log string) string {
	return fmt.Sprintf("INFO: %s", log)
}
