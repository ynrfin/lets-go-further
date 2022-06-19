package main

import (
	"errors"
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
