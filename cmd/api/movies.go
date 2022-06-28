package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ynrfin/greenlight/internal/data"
)

// Add a createMovieHandler for the "POST /v1/movies" endpoint. For now we simply
// return a plain-text placeholder response.
func (app *application) createMovieHandler(w http.ResponseWriter, r *http.Request) {
    // Declare an anonymous struct to held the information that we expect to be in the
    // HTTP request body (note that the field names and types in the struct are a subset
    // of the Movie struct that we created earlier). This struct will be our *target
    // decode destination*
    var input struct {
        Title string `json:"title"`
        Year int32 `json:"year"`
        Runtime data.Runtime `json:"runtime"`
        Genres []string `json:"genres"`
    }

    // Initialize a new json.Decoder instance which reads from the request body, and
    // the use the Decode() method  to decode the body contents into the input struct.
    // Importantly, notice that when we call Decode() we bass a *pointer* to the input
    // struct as the traget decode destination. If there was an error during decoding,
    // we also use our generic errorResponse() helper to send the clien ta 400 bad
    // request response containing the error message.
    err := app.readJSON(w, r, &input)

    if err != nil {
        app.badRequestResponse(w, r, err)
        return
    }
    // Dump the contents of the input struct in a HTTP response.
    fmt.Fprintf(w, "%+v\n", input)
}

// Add a showMovieHandler for the "GET /v1/movies/:id" endpoint. For now, we retrieve
// the interpolated "id" parameter from the current URL and include it in a placeholder
// response.
func (app *application) showMovieHandler(w http.ResponseWriter, r *http.Request) {
    // Declare an anonymous struct to held the information that we expect to be in the
    // HTTP request body (note that the field names and types in the struct are a subset
    // of the Movie struct that we created earlier). This struct will be our *target
    // decode destination*
    var input struct {
        Title string `json:"title"`
        Year int32 `json:"year"`
        Runtime int32 `json:"runtime"`
        Genres []string `json:"genres"`
    }

    // Initialize a new json.Decoder instance which reads from the request body, and
    // the use the Decode() method  to decode the body contents into the input struct.
    // Importantly, notice that when we call Decode() we bass a *pointer* to the input
    // struct as the traget decode destination. If there was an error during decoding,
    // we also use our generic errorResponse() helper to send the clien ta 400 bad
    // request response containing the error message.
    err := json.NewDecoder(r.Body).Decode(&input)

    if err != nil {
        app.errorResponse(w, r, http.StatusBadRequest, err.Error)
        return
    }
    // Dump the contents of the input struct in a HTTP response.
    fmt.Fprintf(w, "%+v\n", input)
}
