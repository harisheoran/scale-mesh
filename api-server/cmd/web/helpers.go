package main

import (
	"fmt"
	"log"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

// Centralized Error Helpers

func (app *app) serverError(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	app.errorLogger.Println(trace)
	log.Println("CHECKING...")
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func (app *app) clientError(g gin.ResponseWriter, status int) {
	http.Error(g, http.StatusText(status), status)
}

func (app *app) notFound(g gin.ResponseWriter) {
	app.clientError(g, http.StatusNotFound)
}
