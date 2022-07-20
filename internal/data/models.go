package data

import (
	"database/sql"
	"errors"
)

// Define a custom ErrRecordNotFound error. We'll return this from our Get() method when
// looking up a movie that does not exists
var (
	ErrRecordNotFound = errors.New("record not found")
)

// Create a Models struct which wraps the MovieModel. We'll add other models to this
// like UserModel and PermissionModel, as our build progress.
type Models struct{
    Movies MovieModel
}

// for ease of use, we also add a New() method which returns a Models struct containing
// the intialized MovieModel
func NewModel(db *sql.DB) Models {
    return Models{
        Movies: MovieModel{DB: db},
    }
}
