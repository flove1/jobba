package main

import (
	"expvar"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)

	router.HandlerFunc(http.MethodGet, "/v1/vacancies", app.requirePermission("vacancies:read", app.listVacanciesHandler))
	router.HandlerFunc(http.MethodGet, "/v1/vacancies/:id", app.requirePermission("vacancies:read", app.showVacancyHandler))
	router.HandlerFunc(http.MethodPost, "/v1/vacancies", app.requirePermission("vacancies:write", app.createVacancyHandler))
	router.HandlerFunc(http.MethodPatch, "/v1/vacancies/:id", app.requirePermission("vacancies:write", app.updateVacancyHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/vacancies/:id", app.requirePermission("vacancies:write", app.deleteVacancyHandler))

	router.HandlerFunc(http.MethodPost, "/v1/users", app.registerUserHandler)
	router.HandlerFunc(http.MethodPut, "/v1/users/activated", app.activateUserHandler)

	router.HandlerFunc(http.MethodPost, "/v1/tokens/authentication", app.createAuthenticationTokenHandler)

	router.Handler(http.MethodGet, "/debug/vars", app.requirePermission("debug", expvar.Handler().ServeHTTP))

	return app.metrics(app.recoverPanic(app.enableCORS(app.rateLimit(app.authenticate(router)))))
}
