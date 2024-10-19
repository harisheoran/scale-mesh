package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/aws/smithy-go"
	"github.com/gin-gonic/gin"
	"gitlab.com/harisheoran/scale-mesh/api-server/pkg/models"
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

func (app *app) deploymentHandler(ctx *gin.Context) {
	deploymentPayload := models.Deployment{}
	err := json.NewDecoder(ctx.Request.Body).Decode(&deploymentPayload)
	if err != nil {
		app.errorLogger.Println("Unable to decode the payload JSON.", err)
		ctx.JSON(http.StatusNotFound, gin.H{
			"response": "Not a valid requeset payload.",
		})
		return
	}

	// second, check the projectID sent is existing id or not
	project, err := app.projectModel.CheckExistingProject(int(deploymentPayload.ProjectID))
	if err == models.ErrNoRecord {
		app.errorLogger.Println("unable to query the project using ID", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"response": "Project Id does not exist",
		})
		return
	} else if err != nil {
		app.errorLogger.Println("unable to query the project using ID", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"response": "Internal Server Error",
		})
		return
	}

	// check if the project belongs to the logged in user
	session, err := app.session.Get(ctx.Request, "thisSession")
	if err != nil {
		app.errorLogger.Println("unable to get the sesstion")
	}
	loggedInUserId := session.Values["id"].(int)

	if uint(loggedInUserId) != project.UserID {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"response": "User not authorized",
		})
		return
	}

	// Check if there is already an existing deployment running or not

	// else

	// configure the AWS SDK to run ECS Task
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("ap-south-1"))
	if err != nil {
		app.errorLogger.Println("Unable to authorise the AWS.", err)
	}
	// Create an ECS Client
	ecsClient := ecs.NewFromConfig(cfg)

	// Run the ECS TASK
	if len(project.GitUrl) != 0 {
		projectID := strconv.FormatUint(uint64(project.ID), 10)

		err = runEcsTask(ecsClient, project.GitUrl, projectID)
		if err != nil {
			app.errorLogger.Println("Unable to Run the ECS TASK", err)
		}

		deploymentPayload.ProjectID = project.ID
		// Save deployment into the Database
		deploymentID, err := app.deploymentController.Insert(deploymentPayload)
		if err != nil {
			app.errorLogger.Println("Unable to save deployment data into DB.", err)
			ctx.JSON(http.StatusInternalServerError,
				gin.H{
					"response": "Internal Server Error.",
				})
		}

		// send the respose to user with website URL
		websiteURL := "http://" + projectID + ".localhost:8080"
		ctx.JSON(http.StatusOK, gin.H{
			"status":        "Start deploying...",
			"websiteUrl":    websiteURL,
			"deployment ID": deploymentID,
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

// /project endpoint to save the project info to db
func (app *app) projectHandler(ctx *gin.Context) {
	projectData := models.Project{}

	body := ctx.Request.Body
	err := json.NewDecoder(body).Decode(&projectData)
	if err != nil {
		app.errorLogger.Println("Unable to parse the decode the JSON payload.")
		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{
				"response": "Internal Server Issue",
			},
		)
		return
	}

	session, err := app.session.Get(ctx.Request, "thisSession")
	currentLoggedInUserID := session.Values["id"].(int)
	if err != nil {
		app.errorLogger.Println("Unable to parse the logged-in user id from cookies to uint.")
		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{
				"response": "Server Issue",
			},
		)
		return
	}

	// set userid key to current logged in user ID (get it from cookie)
	projectData.UserID = uint(currentLoggedInUserID)

	if err != nil {
		app.errorLogger.Println("Unable to decode the request payload JSON.")
		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{
				"response": "Server Issue",
			},
		)
		return
	}

	id, err := app.projectModel.Insert(projectData)
	if err != nil {
		app.errorLogger.Println(err)
		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{
				"response": "Unable to save the data.",
			},
		)
		return
	}

	ctx.JSON(
		http.StatusOK,
		gin.H{
			"response":   "Project info saved successfully.",
			"Project ID": id,
		},
	)
}

// user signup
func (app *app) userSignupHandler(ctx *gin.Context) {
	user := models.User{}

	body := ctx.Request.Body
	err := json.NewDecoder(body).Decode(&user)
	if err != nil {
		app.errorLogger.Println("Unable to Decode the User details", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"response": "Internal Server Error",
		})
		return
	}

	err = app.userDBController.Insert(user)
	if err != nil {
		app.errorLogger.Println("Unable to insert the User details", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"response": "Internal Server Error",
		})
		return
	}

	ctx.JSON(
		http.StatusOK, gin.H{
			"response": "User info saved successfully.",
		},
	)

}

// user login
func (app *app) userLoginHandler(ctx *gin.Context) {
	loginUser := models.LoginUser{}

	err := json.NewDecoder(ctx.Request.Body).Decode(&loginUser)
	if err != nil {
		app.errorLogger.Println("Unable to Decode the User details", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"response": "Internal Server Error..",
		})
		return
	}

	id, err := app.userDBController.Authenticate(loginUser.Email, loginUser.Password)
	if err == models.ErrInvalidCredentials {
		app.errorLogger.Println(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"response": fmt.Sprintf("%s", models.ErrInvalidCredentials),
		})
		return

	} else if err == models.ErrNoRecord {
		app.errorLogger.Println(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"response": fmt.Sprintf("%s", models.ErrNoRecord),
		})
		return
	} else if err != nil {
		app.errorLogger.Println(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"response": "Intenal server error.",
		})
		return
	}

	session, _ := app.session.Get(ctx.Request, "thisSession")
	session.Values["id"] = id
	err = session.Save(ctx.Request, ctx.Writer)
	if err != nil {
		http.Error(ctx.Writer, err.Error(), http.StatusInternalServerError)
		return
	}

	app.infoLogger.Println("saved ID:", id, "COOKIE value:", session.Values["id"])

	ctx.JSON(http.StatusOK, gin.H{
		"response": "User Logged in successfully.",
	})

}

// user logout
func (app *app) userLogoutHandler(ctx *gin.Context) {

	session, _ := app.session.Get(ctx.Request, "thisSession")
	session.Values["id"] = nil
	session.Save(ctx.Request, ctx.Writer)

	log.Println("COOKIE VALUE:", session.Values["id"])

	ctx.JSON(http.StatusOK, gin.H{
		"response": "User Logout successfully.",
	})

}
