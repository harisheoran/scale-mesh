package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/aws/smithy-go"
	"github.com/aws/smithy-go/logging"
	"github.com/gin-gonic/gin"
)

type RepoUrl struct {
	GITHUB_REPO_URL string
}

var baseUrl = "https://scale-mesh-s3.s3.ap-south-1.amazonaws.com/__output/"

func main() {
	logger := logging.NewStandardLogger(os.Stdout)
	// configure the AWS SDK for ECS
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("ap-south-1"),
		config.WithClientLogMode(aws.LogRequestWithBody|aws.LogResponseWithBody), // Log request and response body
		config.WithLogger(logger))
	if err != nil {
		log.Fatal("ERROR: unable to auth with AWS", err)
	}
	ecsClient := ecs.NewFromConfig(cfg)
	fmt.Println(ecsClient)

	// Gin router
	var router = gin.Default()

	router.POST("/project", func(ctx *gin.Context) {
		projectID := generateProjectID(5)
		var githubRepoUrl RepoUrl
		body := ctx.Request.Body

		err := json.NewDecoder(body).Decode(&githubRepoUrl)
		if err != nil {
			ctx.JSON(http.StatusNotFound, gin.H{
				"response": "Send a valid requeset",
			})
			return
		}

		if len(githubRepoUrl.GITHUB_REPO_URL) != 0 {
			// Run the ECS task
			err = runEcsTask(ecsClient, githubRepoUrl.GITHUB_REPO_URL, projectID)
			if err != nil {
				log.Fatal("ERROR: running the task", err.Error())
			}
			websiteURL := "http://" + projectID + ".localhost:8080"
			ctx.JSON(http.StatusOK, gin.H{
				"status":     "Start deploying...",
				"websiteUrl": websiteURL,
			})
		} else {
			ctx.JSON(http.StatusNotFound, gin.H{
				"response": "Send a valid Github Repository URL",
			})
			return
		}
	})

	router.Run(":9000")

}

// Run the ECS task: Run the task from task defination
func runEcsTask(ecsClient *ecs.Client, githubRepoUrlEnv string, projectID string) error {

	cluster := "arn:aws:ecs:ap-south-1:637423604544:cluster/scale-mesh-build"
	taskDefinition := "arn:aws:ecs:ap-south-1:637423604544:task-definition/build-server-container-task:3"
	var count int32 = 1

	// projectID env variable override
	projectIDEnvKey := "projectID"
	projectIDEnvValue := projectID

	// GITHUB_REPO_URL env variable override
	githubRepoUrlEnvKey := "GITHUB_REPO_URL"
	githubRepoUrlEnvValue := githubRepoUrlEnv

	ketValuePairProjectID := types.KeyValuePair{
		Name:  &projectIDEnvKey,
		Value: &projectIDEnvValue,
	}

	keyValuePairGithubRepoUrl := types.KeyValuePair{
		Name:  &githubRepoUrlEnvKey,
		Value: &githubRepoUrlEnvValue,
	}

	containerName := "build-container"
	containerOverride := types.ContainerOverride{
		Environment: []types.KeyValuePair{keyValuePairGithubRepoUrl, ketValuePairProjectID},
		Name:        &containerName,
	}

	taskOverride := types.TaskOverride{
		ContainerOverrides: []types.ContainerOverride{containerOverride},
	}

	awsVpcConfiguration := types.AwsVpcConfiguration{
		Subnets:        []string{"subnet-0cd0b45ca2fe77a01", "subnet-03be364ab5e2e8fb0", "subnet-09d72715eea1c5ab2"},
		SecurityGroups: []string{"sg-088cb654fc20dba4e"},
		AssignPublicIp: types.AssignPublicIpEnabled,
	}

	networkConfiguration := types.NetworkConfiguration{
		AwsvpcConfiguration: &awsVpcConfiguration,
	}

	runTaskInput := ecs.RunTaskInput{
		Cluster:              &cluster,
		TaskDefinition:       &taskDefinition,
		Count:                &count,
		LaunchType:           types.LaunchTypeFargate,
		NetworkConfiguration: &networkConfiguration,
		Overrides:            &taskOverride,
	}

	// Run the TASK
	result, err := ecsClient.RunTask(context.TODO(), &runTaskInput)

	if err != nil {
		var apiErr smithy.APIError // Cast to smithy.APIError to get more detailed error information
		if errors.As(err, &apiErr) {
			fmt.Printf("Error code: %s\n", apiErr.ErrorCode())       // Get the AWS error code
			fmt.Printf("Error message: %s\n", apiErr.ErrorMessage()) // Get the detailed message
		} else {
			fmt.Printf("An unknown error occurred: %v\n", err)
		}
		return err
	}

	fmt.Println("Task launched successfully", result)

	return nil
}

func generateProjectID(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyz0123456789")
	str := make([]rune, n)

	for i := range str {
		str[i] = letters[rand.Intn(len(letters))]
	}

	return string(str)
}
