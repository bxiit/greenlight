package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() *httprouter.Router {
	// Initialize a new httprouter router instance.
	router := httprouter.New()
	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)
	// Register the relevant methods, URL patterns and handler functions for our
	// endpoints using the HandlerFunc() method. Note that http.MethodGet and
	// http.MethodPost are constants which equate to the strings "GET" and "POST"
	// respectively.
	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)
	//router.HandlerFunc(http.MethodPost, "/v1/movies", app.createMovieHandler)
	//router.HandlerFunc(http.MethodGet, "/v1/movies", app.showAllMovieHandler)
	//router.HandlerFunc(http.MethodGet, "/v1/movies/:id", app.showMovieHandler)
	//router.HandlerFunc(http.MethodPut, "/v1/movies/:id", app.updateMovieHandler)
	//router.HandlerFunc(http.MethodDelete, "/v1/movies/:id", app.deleteMovieHandler)

	router.HandlerFunc(http.MethodPost, "/v1/module-infos", app.createModuleInfoHandler)
	router.HandlerFunc(http.MethodGet, "/v1/module-infos/:id", app.getModuleInfoHandler)
	router.HandlerFunc(http.MethodGet, "/v1/module-infos", app.getLatestFiftyModuleInfosHandler)
	router.HandlerFunc(http.MethodPut, "/v1/module-infos/:id", app.editModuleInfoHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/module-infos/:id", app.deleteModuleInfoHandler)
	// Return the httprouter instance.
	return router
}
