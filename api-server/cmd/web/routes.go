package main

import "github.com/gin-gonic/gin"

func (app *app) routes() *gin.Engine {
	router := gin.Default()
	router.GET("/health", app.healthHandler)
	router.POST("/project", app.runEcsTaskHandler)

	return router
}
