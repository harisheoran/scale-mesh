## Scale Mesh
Deploy your web apps with ease.

## Architecture
![](./img/arch.png)

## Setup and run locally.
### Requirements
- Go version *go1.23.2*
- Install *air* package for live reload
    ```
    go install github.com/air-verse/air@latest
    ```

- Clone the repository.

1. Run the *api-server* (main backend server)
```
cd api-server
```
Install the dependencies

```
go mod download
```

Run the Postgre SQL database
```
docker run --name pg-docker -e POSTGRES_USER=admin -e POSTGRES_PASSWORD=secret -e POSTGRES_DB=mydb -p 5432:5432 -d postgres
```
Export the Database URI

```
export DBURI="host=localhost user=admin password=secret dbname=mydb port=5432 sslmode=disable TimeZone=UTC"
```

Export the Session secret
```
export SESSION_KEY=mysecret
```

Run the API
```
air --build.cmd "go build -o bin/api ./cmd/web/" --build.bin "./bin/api"
```

this API will start at port 9000, check the api at ```127.0.0.1:9000/health```.

2. Run the reverse-proxy API

```
cd reverse-proxy
go mod download
air --build.cmd "go build -o bin/api ./cmd/web/main.go" --build.bin "./bin/api"
```
3. Build the *build-server* container
Go to the Build server directory
```
cd build-server
```

Build the image
```
docker build -t scale-mesh/build-server-container-image .
```

Run the container, it require 2 env variables.

```GITHUB_REPO_URL```
```projectID```

## Components
1. ***Build Server***

2. ***Reverse Proxy API***

3. ***Main API Server***

4. ***Frontend Server***

5. ***Log Collection Pipeline***


## Build Server
To build the code and push the artifacts to the S3 bucket.

### How it works?
It is a custom Docker container which uses the GitHub repo url, clones it and then build it and push them to SS3 bucket

***Structure***
- Dockerfile
- entry.sh
- main.go

> It is a Multistage Docker build, which builds the Go binary in first stage and then run the bianry in 2nd stage to build the user's web app code and push the build artifacts to the S3 bucket.

Build Server image is pushed to AWS ECR, and then a ECS cluster & Task defination are created to run a container from the ECR image, and after task completed, it'll destroy the container.

> S3 buckets are uploaded in this path
***__output/{projectID}/*** 

## API Server
Main backend API, via which user interacts with the Web app.

### Working Flow
- User Signup/login 
- Save the user's web app project info(GitHub repo url, name) to deploy.
- Deploy the web app.


### Structure 
- ***cmd/web/***
Contains all the API and business logic.

- ***pkg***
Contains the ancillary non-application-specfic code, which could potentially be reused.

- Contains the Database models and methods.

### Endpoints of API server
- ***GET*** ```/health``` to check health of the API.
- ***POST*** ```/deploy``` to deploy the project.
- ***POST*** ```/project``` to save the info of the project.
- ***POST*** ```/user/signup``` to signup.
- ***POST*** ```/user/login``` to login.
- ***POST*** ```/user/logout``` to logout.

## Reverse Proxy API
To serve the user Web App dynamically using the unique project id.

## Frontend Server
Serve a basic HTML template for user to interact with the application.

## Logging Pipeline
Build-Server container pushes the logs to the Redis using pub/sub feature and a web socker server is subscribing to the redis channel.


## Application Deployment (DevOps)

### CI Pipeline (GitLab CICD)
- To build and push the image to Docker Hub registry of API server.

### Deployment on Kubernetes
- ```k8s``` directory contains the YAML files to deploy the web app on k8s.

- How to deploy
    - Create Namespace.
    - Create secret
    - Create DB
    - Deploy service.

