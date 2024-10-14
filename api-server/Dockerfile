FROM golang:1.23.1 as build

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/bin/api-server ./cmd/web/

FROM alpine:3.20 as prod

WORKDIR /app

COPY --from=build /app/bin/api-server /app/bin/api-server

RUN chmod +x /app/bin/api-server

EXPOSE 9000

ENTRYPOINT [ "/app/bin/api-server" ]