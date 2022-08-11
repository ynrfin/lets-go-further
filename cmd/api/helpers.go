package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

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
	// Use http.maxBytesReader() to limit the size of the request body to 1MB.
	maxBytes := 1_048_576
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	// Initialize the json.Decoder, and call the DissallowUnknownFields() method on it
	// before decoding. This means that if the JSON from the client now includes any
	// field which cannot be mapped to the target destination, the decoder will return
	// and error instaed of just ignoring the field.
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	err := dec.Decode(dst)
	if err != nil {
		// if there is an error during decoding, start the triage
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError

		switch {
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

		// If the JSON contains a field which cannot be mapped tothe target destination
		// then Decode() wil now return an error message in the fromat "json: unknown
		// feild "<name>"". We check for this, extract the field name from the error,
		// and interpolate it into ur custom error message. Note that there's an open
		// issuse at htps://github.com/golang.go/issues/29035 regarding turning this
		// into distinct error type in the future
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return fmt.Errorf("body contains unknown key %s", fieldName)

		// If the request body exceeds 1MB in size the decode will now fail with the
		// error "http: request body too large". There is an open issu about turning
		// tis into distinct error type at https://github.com/golagn/go/issuse/30175
		case err.Error() == "http: request body too large":
			return fmt.Errorf("body must not be larger than %d bytes", maxBytes)

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

	// Call Decode() again, using pointer to an empty anonymous struct as the
	// destination. If the requestbody only contained a sigle JSON value this will
	// return an io.EOF error. So if we get anything else, we know that there is
	// additional data in the request body and we return our custom error message.
	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must only contain a single JSON value")
	}
	return nil
}

// The background() helper accepts an arbitrary function as a parameter.
func (app *application) background(fn func()) {
	// Use defer to decrement the WaitGroup counter before the goroutine returns.
	defer app.wg.Done()

	// Launch a background goroutine
	go func() {
		// recover any panic
		defer func() {
			if err := recover(); err != nil {
				app.logger.PrintErr(fmt.Errorf("%s", err), nil)
			}
		}()

		// execute the arbitrary functio that we passed as the parameter.
		fn()
	}()
}
