package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/lib/pq"
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

func ValidateMovie(v *validator.Validator, movie *Movie) {

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
	v.Check(len(movie.Genres) >= 1, "genres", "must contain at least 1 genre")
	v.Check(len(movie.Genres) <= 5, "genres", "must not contain more than 5 genre")

	v.Check(validator.Unique(movie.Genres), "genres", "must not contain duplicate values")

}

type MovieModel struct {
	DB *sql.DB
}

// Add placeholder method for inserting a new record in the movies table
func (m MovieModel) Insert(movie *Movie) error {
	// Define the sql query for inserting a new record in the movies table and returning
	// the system-generated data
	query := `
        INSERT INTO movies (title, year, runtime, genres)
        VALUES ($1, $2, $3, $4)
        RETURNING id, created_at, version `

	// create an args slice containing the values for the placeholder parameters from
	// the movie struct. Declaring this slice immeditely next to ou SQL query helps to
	// make it nice and clear *what values are being used where* in the query
	args := []any{movie.Title, movie.Year, movie.Runtime, pq.Array(movie.Genres)}

	// Create a context wit 3-second timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Use the QueryRow() method to execute the SQL query on our connection pool,
	// passing in the args slice as a variadic parameter and scanning the system-
	// generated id, created_at and version values into the movie struct.
	return m.DB.QueryRowContext(ctx, query, args...).Scan(&movie.ID, &movie.CreatedAt, &movie.Version)
}

// Add placeholder method for fetching a specific record from the movies table
func (m MovieModel) Get(id int64) (*Movie, error) {
	// The PostgreSQL bigserial type that we're using for the movie  ID starts
	// auto-incrementing at 1 by default, so we know that no movies will ahve ID values
	// less than that. TO avoid making an unnescessary database call, we take a shortcut
	// and return an ErrRecordNotFound error straight away.
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	// Define the SQL query for retrieving the movie data.
	query := `
        SELECT  id, created_at, title, year, runtime, genres, version
        FROM movies
        WHERE id = $1 `

	var movie Movie

	// Use the context.WithTimeout() function to create a contet.Context which carries a
	// 3-second timeout deadiline. Note that we're using the empty context.Background()
	// as the 'parent' context
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	// Importantly, use devefr to make sure that we cancel the context before the Get()
	// method returns.
	defer cancel()

	// Execute the query using the QueryRow() method, passing in the provided id value
	// as a placeholder parameter, and scan the response data into the fields of the
	// Movie struct. Importantly, notice that we need to convert the scan target for the
	// genres column using the pq.Array() adapter function again.
	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&movie.ID,
		&movie.CreatedAt,
		&movie.Title,
		&movie.Year,
		&movie.Runtime,
		pq.Array(&movie.Genres),
		&movie.Version,
	)

	// Handle any errors. If there was no matching movie found, Scan() will return
	// a sql.ErrNoRows error. We check for this and return our custom ErrRecordNotFound
	// error instead.
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &movie, nil
}

// Add placeholder method for updating a specific record from the movies table
func (m MovieModel) Update(movie *Movie) error {
	// Declare the SQL query for updating the record and returning the new version
	// number
	query := `
        UPDATE movies
        SET title = $1, year = $2, runtime= $3, genres = $4, version = version + 1
        WHERE id = $5 AND version = $6
        RETURNING version
    `

	// Create an args slice containing the values fro the placeholder parameters.
	args := []any{
		movie.Title,
		movie.Year,
		movie.Runtime,
		pq.Array(movie.Genres),
		movie.ID,
		movie.Version,
	}

	// Create a context wit 3-second timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// use the QueryRow() method to execute the qery, passing in the args slice as a
	// variadic parameter and scanning the new version value into the movie struct
	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&movie.Version)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}
	return nil
}

func (m MovieModel) Delete(id int64) error {
	// Return an ErrRecordNotFound error if the movie ID is less than 1.
	if id < 1 {
		return ErrRecordNotFound
	}

	// Construct the SQL query to delete the record.
	query := `
        DELETE FROM movies
        where id = $1
    `

	// Create a context wit 3-second timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Execute the SQL query using the Exec() method, passing in the id variable as
	// the value for the placeholder parameter. The Exec() method erturs a sql.Result
	// object.
	result, err := m.DB.ExecContext(ctx,query, id)
	if err != nil {
		return err
	}

	// Call the RowsAffected() method on the sql.Result object to ge the number of rows
	// affected by the query.
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	// If no rows were affected, we know that the movies table didn't contain a record
	// with the provided ID at the moment we tried to delete it. In that case we
	// return an ErrRecordNotFound error.
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}
	return nil
}
