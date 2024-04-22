package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// todo *httprouter.Router
func (app *application) routes() http.Handler {
	// Initialize a new httprouter router instance.
	router := httprouter.New()
	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)
	// Register the relevant methods, URL patterns and handler functions for our
	// endpoints using the HandlerFunc() method. Note that http.MethodGet and
	// http.MethodPost are constants which equate to the strings "GET" and "POST"
	// respectively.
	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)

	// user-info
	router.HandlerFunc(http.MethodPost, "/v1/user-infos", app.registerUserInfoHandler)                             // register
	router.HandlerFunc(http.MethodPost, "/v1/user-infos/activated", app.activateUserInfoHandler)                   // activate
	router.HandlerFunc(http.MethodPost, "/v1/tokens/authentication", app.createAuthenticationTokenHandlerUserInfo) // authenticate
	router.HandlerFunc(http.MethodGet, "/v1/user-infos/:id", app.requireActivatedUserInfo(app.getUserInfoHandler)) // get
	router.HandlerFunc(http.MethodGet, "/v1/user-infos", app.requireActivatedUserInfo(app.getAllUserInfoHandler))  // getAll
	router.HandlerFunc(http.MethodPut, "/v1/user-infos/:id", app.requireAdminRole(app.editUserInfoHandler))        // edit
	router.HandlerFunc(http.MethodDelete, "/v1/user-infos/:id", app.requireAdminRole(app.deleteUserInfoHandler))   // delete

	// movie
	//router.HandlerFunc(http.MethodPost, "/v1/movies", app.createMovieHandler)
	//router.HandlerFunc(http.MethodGet, "/v1/movies", app.showAllMovieHandler)
	//router.HandlerFunc(http.MethodGet, "/v1/movies/:id", app.showMovieHandler)
	//router.HandlerFunc(http.MethodPut, "/v1/movies/:id", app.updateMovieHandler)
	//router.HandlerFunc(http.MethodDelete, "/v1/movies/:id", app.deleteMovieHandler)

	router.HandlerFunc(http.MethodPost, "/v1/module-infos", app.requireAdminRole(app.createModuleInfoHandler))
	router.HandlerFunc(http.MethodGet, "/v1/module-infos/:id", app.requireActivatedUserInfo(app.getModuleInfoHandler))
	router.HandlerFunc(http.MethodGet, "/v1/module-infos", app.requireActivatedUserInfo(app.getLatestFiftyModuleInfosHandler))
	router.HandlerFunc(http.MethodPut, "/v1/module-infos/:id", app.requireAdminRole(app.editModuleInfoHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/module-infos/:id", app.requireAdminRole(app.deleteModuleInfoHandler))

	// dep
	router.HandlerFunc(http.MethodPost, "/v1/department-infos", app.requireAdminRole(app.createDepInfoHandler))
	router.HandlerFunc(http.MethodGet, "/v1/department-infos/:id", app.requireActivatedUserInfo(app.getDepartmentInfoHandler))

	// users
	//router.HandlerFunc(http.MethodPost, "/v1/users", app.registerUserHandler)
	//router.HandlerFunc(http.MethodPut, "/v1/users/activated", app.activateUserHandler)
	// Add the route for the POST /v1/tokens/authentication endpoint.
	//router.HandlerFunc(http.MethodPost, "/v1/tokens/authentication", app.createAuthenticationTokenHandler)
	// Return the httprouter instance.
	return app.recoverPanic(app.rateLimit(app.authenticate(router)))
}
