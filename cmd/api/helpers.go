package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

// Retrieve  the "id" URL parameterfrom the current request context, then convert it to
// an integer and return it. If the operation isn't successful, return 0 and and error mesage
func (app *application) readIdParam(r *http.Request) (int64, error) {
	params := httprouter.ParamsFromContext(r.Context())

	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, errors.New("invalid id parameter")
	}
	return id, nil
}

type envelope map[string]any

// Define a writeJSON() helper for sending responses. This takes the destination
// http.ResponseWriter, the HTTP status code to send, the data to encode to JSON, and a
// header map containing any additional HTTP headers we want to include in the response.
func (app *application) writeJSON(w http.ResponseWriter, status int, data any, headers http.Header) error {
	// Encode the data to JSON, returning the error if there was one
	js, err := json.MarshalIndent(data, "", "\t")

	if err != nil {
		return err
	}

	// Append a newline to make ieasier to view in terminal application
	js = append(js, '\n')

	// At this point, we know that we won't encounter any more errors before wrting the
	// response, so it's safe to add any headers that we want to include. We loop
	// through the header map and add each haeder to the http.ResponseWriter header map.
	// Note that it's OK if the provided header map is nil. GO doesn't throw an error
	// if you try to range over(or generally, read from) a nil map.
	for key, value := range headers {
		w.Header()[key] = value
	}

	// Add the "Content-Type: application/json" header, then write the status code and
	// JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)
	return nil
}

func (app *application) readJSON(w http.ResponseWriter, r *http.Request, dst any) error {
    // Decode the request body into the traget destination
    err := json.NewDecoder(r.Body).Decode(dst)

    if err != nil {
        // if there is an error during decoding, start the triage
        var syntaxError *json.SyntaxError
        var unmarshalTypeError *json.UnmarshalTypeError
        var invalidUnmarshalError *json.InvalidUnmarshalError

        switch{
        // Use the errors.As() function tocheck wether the error has the type
        // *json.SyntaxError. If it does, then return a plain english error message
        // which includes the loction of the problem
        case errors.As(err, &syntaxError):
            return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)

        // In some circumstances Decode() may also return and io.ErrUnexpectedEOF error
        // for syntax errrors in the JSON. So we check for this using errors.Is() and
        // return  generic error message. THere is an open issue regarding this at
        // github
        case errors.Is(err, io.ErrUnexpectedEOF):
            return errors.New("body contains badly formed JSON")

        // likewise, catch any *json.UnmarshalTypeError errors. These occur when the
        // JSON value is the wrong type for the target destination. If the error relates
        // to a specific field, then we include that in ourerror messageto make it
        // easier for the client to debug
        case errors.As(err, &unmarshalTypeError):
            if unmarshalTypeError.Field != "" {
                return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
            }
            return fmt.Errorf("body contains incorrect JSON type(at character %d)", unmarshalTypeError.Offset)

        // A io.EOF erro will be returned by Decode() if the ruqest body is empty. WE
        // check for this with errrors.Is() and return a plain-englis eror message
        // intaed
        case errors.Is(err, io.EOF):
            return errors.New("body must not be empty")

        // A json.InvalidUnmarshalError error will bereturned if we pass
        // something that is not a non-nil pointer to Decode(). We cathc this and panic
        // rather than returning an error to our handler. At the end of this chapter
        /// we'll talk about panicking versus returning errors, and discurss why it's an
        // appropriate thing to do in this specific situation
        case errors.As(err, &invalidUnmarshalError):
            panic(err)

        default:
            return err
        }
    }
    return nil
}
