package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// todo *httprouter.Router
func (app *application) routes() http.Handler {
	router := httprouter.New()
	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)
	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)

	// user-info
	router.HandlerFunc(http.MethodPost, "/v1/user-infos", app.RegisterUserInfoHandler)                             // register
	router.HandlerFunc(http.MethodPost, "/v1/user-infos/activated", app.ActivateUserInfoHandler)                   // activate
	router.HandlerFunc(http.MethodPost, "/v1/tokens/authentication", app.CreateAuthenticationTokenHandlerUserInfo) // authenticate
	router.HandlerFunc(http.MethodGet, "/v1/user-infos/:id", app.requireActivatedUserInfo(app.GetUserInfoHandler)) // get
	router.HandlerFunc(http.MethodGet, "/v1/user-infos", app.requireActivatedUserInfo(app.GetAllUserInfoHandler))  // getAll
	router.HandlerFunc(http.MethodPut, "/v1/user-infos/:id", app.requireAdminRole(app.EditUserInfoHandler))        // edit
	router.HandlerFunc(http.MethodDelete, "/v1/user-infos/:id", app.requireAdminRole(app.DeleteUserInfoHandler))   // delete

	//module-info
	router.HandlerFunc(http.MethodPost, "/v1/module-infos", app.requireAdminRole(app.CreateModuleInfoHandler))
	router.HandlerFunc(http.MethodGet, "/v1/module-infos/:id", app.requireActivatedUserInfo(app.GetModuleInfoHandler))
	router.HandlerFunc(http.MethodGet, "/v1/module-infos", app.requireActivatedUserInfo(app.GetLatestFiftyModuleInfosHandler))
	router.HandlerFunc(http.MethodPut, "/v1/module-infos/:id", app.requireAdminRole(app.EditModuleInfoHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/module-infos/:id", app.requireAdminRole(app.DeleteModuleInfoHandler))

	// dep
	router.HandlerFunc(http.MethodPost, "/v1/department-infos", app.requireAdminRole(app.createDepInfoHandler))
	router.HandlerFunc(http.MethodGet, "/v1/department-infos/:id", app.requireActivatedUserInfo(app.getDepartmentInfoHandler))

	// users
	//router.HandlerFunc(http.MethodPost, "/v1/users", App.registerUserHandler)
	//router.HandlerFunc(http.MethodPut, "/v1/users/activated", App.activateUserHandler)
	// Add the route for the POST /v1/tokens/authentication endpoint.
	//router.HandlerFunc(http.MethodPost, "/v1/tokens/authentication", App.CreateAuthenticationTokenHandler)
	// Return the httprouter instance.
	return app.recoverPanic(app.rateLimit(app.authenticate(router)))
}
