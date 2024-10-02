# Build stage: build the golang binary
FROM golang:1.23.1 as buildscript

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY *.go ./

#RUN go build -o /app/bin/build-server main.go (statically compile it)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/bin/build-server main.go

RUN chmod +x /app/bin/build-server

#################
# Prod stage
#################
FROM ubuntu:focal as prod

WORKDIR /app

COPY --from=buildscript /app/bin/build-server /app/bin/build-server

RUN apt-get update && apt-get install -y curl git wget

# Install Node JS and npm
RUN curl -fsSL https://deb.nodesource.com/setup_20.x | bash - \
&& apt-get install -y nodejs

# copy the script file and set executable permissions
COPY entry.sh .
RUN chmod +x /app/entry.sh /app/bin/build-server

ENTRYPOINT ["/app/entry.sh"]