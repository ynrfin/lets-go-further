package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() *httprouter.Router {
	// initialize a new httprouter instance.
	router := httprouter.New()

	// Register the relevant methods, URL patterns and handler functions for our
	// endpoints using the HandleFunc() method. Note that http.MethodGet and
	// http.MethodPost are constants which equate to the stirngs "GET" and "POST"
	// respectively.
	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)
	router.HandlerFunc(http.MethodPost, "/v1/movies", app.createMovieHandler)
	router.HandlerFunc(http.MethodGet, "/v1/movies/:id", app.showMovieHandler)

	return router

}
