package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"math/rand"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/aws/smithy-go"
	"github.com/gin-gonic/gin"
)

// Health Check endpoint handler
func (app *app) healthHandler(context *gin.Context) {
	context.JSON(
		http.StatusOK,
		gin.H{
			"response": "API Health is Ok.",
		},
	)
}

func (app *app) runEcsTaskHandler(ctx *gin.Context) {
	projectID := generateProjectID(5)

	// configure the AWS SDK for ECS
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("ap-south-1"))
	if err != nil {
		app.errorLogger.Println("Unable to authorise the AWS.", err)
	}
	// Create an ECS Client
	ecsClient := ecs.NewFromConfig(cfg)

	// object to decode the payload JSON
	var githubRepoUrl RepoUrl

	// Decode the payload JSON
	body := ctx.Request.Body
	err = json.NewDecoder(body).Decode(&githubRepoUrl)
	if err != nil {
		app.errorLogger.Println("Unable to decode the payload JSON.", err)
		ctx.JSON(http.StatusNotFound, gin.H{
			"response": "Send a valid requeset",
		})
		return
	}

	// Run the ECS TASK
	if len(githubRepoUrl.GITHUB_REPO_URL) != 0 {
		err = runEcsTask(ecsClient, githubRepoUrl.GITHUB_REPO_URL, projectID)
		if err != nil {
			app.errorLogger.Println("Unable to Run the ECS TASK", err)
		}
		// send the respose to user with website URL
		websiteURL := "http://" + projectID + ".localhost:8080"
		ctx.JSON(http.StatusOK, gin.H{
			"status":     "Start deploying...",
			"websiteUrl": websiteURL,
		})
	} else {
		app.errorLogger.Println("Payload GitHub URL is not valid.")
		ctx.JSON(http.StatusNotFound, gin.H{
			"response": "Send a valid Github Repository URL",
		})
		return
	}

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
	_, err := ecsClient.RunTask(context.TODO(), &runTaskInput)

	if err != nil {
		var apiErr smithy.APIError // Cast to smithy.APIError to get more detailed error information
		if errors.As(err, &apiErr) {
			log.Printf("ERROR: code %s\n", apiErr.ErrorCode())       // Get the AWS error code
			log.Printf("ERROR: message %s\n", apiErr.ErrorMessage()) // Get the detailed message
		} else {
			log.Printf("An unknown error occurred: %v\n", err)
		}
		return err
	}

	return nil
}

// Generate a random Project ID
func generateProjectID(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyz0123456789")
	str := make([]rune, n)

	for i := range str {
		str[i] = letters[rand.Intn(len(letters))]
	}

	return string(str)
}

// websocket handeler
