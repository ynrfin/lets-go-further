package data

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/base32"
	"time"

	"github.com/ynrfin/greenlight/internal/validator"
)

// Define constants for the token scope. For now we justdefine the scope "activation"
// but we'll add additional scopes later in the book.
const (
	ScopeActivation = "activation"
)

// Define a Token struct to hold the data for an individual token. This includes the
// plaintext and hashed versions of the token., associated user ID, expiry time and
// scope.
type Token struct {
	Plaintext string
	Hash      []byte
	UserID    int64
	Expiry    time.Time
	Scope     string
}

func generateToken(userID int64, ttl time.Duration, scope string) (*Token, error) {
	// Create a Token instance containig the user ID, expiry, and scope information.
	// Notice that we did add the provided ttl (time-to-live) duration paramter to the
	// current time to get the expiry time?
	token := &Token{
		UserID: userID,
		Expiry: time.Now().Add(ttl),
		Scope:  scope,
	}

	// Initialize a zero-valued byte stlice with a length of 16 bytes
	randomBytes := make([]byte, 16)

	// Encode the byte slice to a base 32 encoded string and assign it to the token
	// Plaintext field. This wiil be the token string that we send to the user in their
	// welcome email, Theyw will look similar to this:
	//
	// ENHINDEDULJl39jqj8lhEHINI
	//
	// Note that by default base 32 string may be padded at the end wit the =
	// character. We dont need this padding charracter for the purpose of our tokens, so
	// we can use the WithPadding(base32.StdEncoding) method in the line below to omit them.
	token.Plaintext = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)

	// Generate a SHA-256 of the plaintext token string. THis will be the value
	// that we store in the `hash` field of our database table. Note that the
	// sha256.Sum256() function returns an *array* of length 32, so to make it easier to
	// work with we convert it to a slice using the [:] operator before storing it.
	hash := sha256.Sum256([]byte(token.Plaintext))
	token.Hash = hash[:]

	return token, nil
}

// Check that the plaintext token has bn provided and is exactly 26 bytes long
func ValidateTokenPlaintext(v *validator.Validator, tokenPlaintext string) {
	v.Check(tokenPlaintext != "", "token", "must be provided")
	v.Check(len(tokenPlaintext) == 26, "token", "must be 26 bytes long")
}

// Define the TokenModel type.
type TokenModel struct {
	DB *sql.DB
}

// The New()m method is a shortcut which creates a new Token struct and then inserts the
// data in the tokens table.
func (m TokenModel) New(userID int64, ttl time.Duration, scope string) (*Token, error) {
	token, err := generateToken(userID, ttl, scope)
	if err != nil {
		return nil, err
	}
	err = m.Insert(token)
	return token, err
}

// Insert() adds the data for a specific token to the tokens table.
func (m TokenModel) Insert(token *Token) error {
	query := `
    INSERT INTO tokens (hash, user_id, expiry, scope)
    VALUES ($1, $2, $3, $4)
    `
	args := []any{token.Hash, token.UserID, token.Expiry, token.Scope}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := m.DB.ExecContext(ctx, query, args...)
	return err
}

// DeleteAllForUser() deletes all token for a specific user and scope
func (m TokenModel) DeleteAllForUser(scope string, userID int64) error {
	query := `
        DELETE FROM tokens
        WHERE scope = $1 AND users_id = $2
    `
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := m.DB.ExecContext(ctx, query, scope, userID)
	return err
}

