package data

import (
	"time"

	"github.com/ynrfin/greenlight/internal/validator"
)

type Movie struct {
	ID        int64     `json:"id"`                // Unique integer ID for the movie
	CreatedAt time.Time `json:"-"`                 // Timestam for when the movie is added to our database
	Title     string    `json:"title"`             // Movie title
	Year      int32     `json:"year,omitempty"`    // Movie release year
	Runtime   Runtime   `json:"runtime,omitempty"` //Movie runtime (in minutes)
	Genres    []string  `json:"genres,omitempty"`  // Slice of genres for the movie
	Version   int32     `json:"version"`           // The version number starts at 1 and will be incremented each time when the movie information is updated
}

func ValidateMovie(v *validator.Validator, movie *Movie){

    // Use Check() method to execute our validation checks. This will add the
    // provided key and error message to the errors mp if the check does not evaluate
    // to true.
    v.Check(movie.Title != "", "title", "must be provided")
    v.Check(len(movie.Title) <= 500, "title", "must not be more than 500 bytes long")

    v.Check(movie.Year != 0, "year", "must be provided")
    v.Check(movie.Year >= 1888, "year", "must be greater than 1888")
    v.Check(movie.Year <= int32(time.Now().Year()), "year", "must not be in the future")

    v.Check(movie.Runtime != 0, "runtime", "must be provided")
    v.Check(movie.Runtime > 0, "runtime", "must be positive integer")

    v.Check(movie.Genres != nil, "genres", "must be provided")
    v.Check(len(movie.Genres) >= 1 , "genres", "must contain at least 1 genre")
    v.Check(len(movie.Genres) <= 5 , "genres", "must not contain more than 5 genre")

    v.Check(validator.Unique(movie.Genres), "genres", "must not contain duplicate values")

}
