package main

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func main() {

	var projectID = os.Getenv("project_id")

	// Authenticate with AWS SDK
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("ap-south-1"))
	if err != nil {
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
		log.Fatal("ERROR: unable to find cloned repo", err)
	}

	// build the application
	log.Println("Building the application ...")
	cmd := exec.Command("/bin/sh", "-c", "npm install && npm run build")

	_, err = cmd.CombinedOutput()
	if err != nil {
		log.Fatal("ERROR: unable to build the application", err)
	}
	log.Println("Build completed...")

	/*
	   2. Upload the build artifacts to S3 buckets.
	*/

	log.Println("Uploading build artifacts to S3 bucket...")
	// change the directory to the build output directory
	buildOutputPath := "/app/output/dist/"
	err = os.Chdir(buildOutputPath)
	if err != nil {
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
					log.Fatal("ERROR: uploading the file to S3", err)
				}
			}
			return nil
		},
	)
	if err != nil {
		fmt.Println(err)
	}
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

	contentType := filepath.Ext(filename)

	_, err = s3Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String(bucketName),
		Body:        file,
		Key:         &objectKey,
		ContentType: &contentType,
	})

	if err != nil {
		return err
	}

	return nil
}
