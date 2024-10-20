package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (app *app) routes() http.Handler {

	router := gin.Default()

	router.GET("/health", app.healthHandler)
	router.POST("/deploy", app.requireAuthenticatedUserMiddleware(app.deploymentHandler))
	router.POST("/project", app.requireAuthenticatedUserMiddleware(app.projectHandler))

	router.POST("/user/signup", app.userSignupHandler)
	router.POST("/user/login", app.userLoginHandler)
	router.POST("/user/logout", app.userLogoutHandler)

	return app.recoverPanic((secureHeaderMiddleware(router)))
}
