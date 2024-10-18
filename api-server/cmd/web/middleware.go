package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func secureHeaderMiddleware(next http.Handler) http.Handler {
	newHandler := func(w http.ResponseWriter, request *http.Request) {
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("X-Frame-Options", "deny")

		// call the next handler
		next.ServeHTTP(w, request)
	}
	return http.HandlerFunc(newHandler)
}

func (app *app) logRequestMiddleware(next http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, request *http.Request) {
		app.infoLogger.Printf("%s - %s %s %s", request.RemoteAddr, request.Proto, request.Method)

		next.ServeHTTP(w, request)
	}

	return http.HandlerFunc(handler)
}

func (app *app) recoverPanic(next http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, request *http.Request) {
		// create a defer function which always run in case of panic as go unwinds the stack
		defer func() {
			// use the builtin recover function to check if there has been a panic or not.
			if err := recover(); err != nil {
				// set header to close
				// Close header on the response acts as a
				//trigger to make Go’s HTTP server automatically close the current
				//connection after a response has been sent.
				w.Header().Set("Connection", "close")
				app.serverError(w, fmt.Errorf("%s", err))
			}
		}()
		next.ServeHTTP(w, request)
	}

	return http.HandlerFunc(handler)
}

func (app *app) requireAuthenticatedUserMiddleware(next gin.HandlerFunc) gin.HandlerFunc {
	handler := func(ctx *gin.Context) {
		session, _ := app.session.Get(ctx.Request, "thisSession")
		app.infoLogger.Println("Middleware hit")
		app.infoLogger.Println("Current cookie Value", session.Values["id"])
		if session.Values["id"] == nil {
			app.infoLogger.Println("COOKIE VALUE checked:", session.Values["id"])
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"response": "User is not authorized to do this.",
			})
			return
		}

		next(ctx)
	}

	return handler
}
