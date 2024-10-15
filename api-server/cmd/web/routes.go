package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (app *app) routes() http.Handler {
	router := gin.Default()
	router.GET("/health", app.healthHandler)
	router.POST("/deploy", app.deploymentHandler)
	router.POST("/project", app.projectHandler)

	return secureHeaderMiddleware(router)
}
