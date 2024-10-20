package main

import (
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func (app *app) routes() http.Handler {

	router := gin.Default()
	// Apply CORS middleware with configuration
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://your-frontend.com"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,           // Enable cookies and credentials
		MaxAge:           12 * time.Hour, // Cache preflight response
	}))

	router.GET("/health", app.healthHandler)
	router.POST("/deploy", app.requireAuthenticatedUserMiddleware(app.deploymentHandler))
	router.POST("/project", app.requireAuthenticatedUserMiddleware(app.projectHandler))

	router.POST("/user/signup", app.userSignupHandler)
	router.POST("/user/login", app.userLoginHandler)
	router.POST("/user/logout", app.userLogoutHandler)

	return app.recoverPanic((secureHeaderMiddleware(router)))
}
