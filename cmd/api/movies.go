package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ynrfin/greenlight/internal/data"
	"github.com/ynrfin/greenlight/internal/validator"
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

    err := app.readJSON(w, r, &input)
    if err != nil {
        app.badRequestResponse(w, r, err)
    }

    // Note that the mevoe variable contains a *pointer* to a Movie struct
    movie := &data.Movie{
        Title: input.Title,
        Year: input.Year,
        Runtime: input.Runtime,
        Genres: input.Genres,
    }

    v := validator.New()

    if data.ValidateMovie(v, movie);!v.Valid() {
        app.failedValidationResponse(w, r, v.Errors)
        return
    }

    // Call the Insert() method on our movies model, passing in a pointer to the
    // validated movie struct. This will create a record in the database and update the
    // movie struct with the system-genrated information.
    err = app.models.Movies.Insert(movie)
    if err != nil {
        app.serverErrorResponse(w, r, err)
        return
    }

    // When sending a HTTP respones, we want to include a Location header to let the
    // client know which URL they can find the newly-created resource at. We make an
    // empty http.Header map and then use the Set() method to add a new Location header.
    // interpolating the system-generated ID for our new movie in the URL.
    headers := make(http.Header)
    headers.Set("Location", fmt.Sprintf("/v1/movies/%d", movie.ID))

    // Write a JSON response with 201 Created status code. the movie data in the
    // response body, and the Location header.
    err = app.writeJSON(w, http.StatusCreated, envelope{"movie": movie}, headers)
    if err != nil{
        app.serverErrorResponse(w, r, err)
    }
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
