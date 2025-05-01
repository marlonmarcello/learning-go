package models

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID             int
	Name           string
	Email          string
	HashedPassword []byte
	Created        time.Time
}

type UserModel struct {
	DB  *pgxpool.Pool
	CTX context.Context
}

func (m *UserModel) Insert(name, email, password string) (int, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		// Wrap the error for context
		return 0, fmt.Errorf("failed to hash password: %w", err)
	}

	// SQL statement to insert user data and return the generated ID.
	// Uses CURRENT_TIMESTAMP for the 'created' field directly in the database.
	statement := `INSERT INTO users (name, email, hashed_password, created)
	              VALUES($1, $2, $3, CURRENT_TIMESTAMP)
	              RETURNING id`

	var newId int

	// Execute the query using QueryRow (since we expect one row back with RETURNING id)
	// and scan the returned ID into the newId variable.
	// Ensure m.CTX provides a valid context.
	err = m.DB.QueryRow(m.CTX, statement, name, email, string(hashedPassword)).Scan(&newId)
	if err != nil {
		// Check if the error is specifically a PostgreSQL error.
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			// Check if the error code corresponds to 'unique_violation'.
			// The SQLSTATE code for unique_violation is "23505".
			// Also check if the violation occurred on the correct constraint.
			// Replace "users_email_key" with your actual constraint name if it differs.
			// You can find the constraint name using `\d users` in psql.
			if pgErr.Code == "23505" && pgErr.ConstraintName == "users_email_key" {
				// If it's a unique violation on the email constraint, return our specific error.
				return 0, ErrDuplicateEmail
			}
		}
		return 0, err
	}

	// If no error occurred, return the ID of the newly created user.
	return newId, nil
}

func (m *UserModel) Authenticate(email, password string) (int, error) {
	var id int
	var hashedPassword []byte

	statement := "SELECT id, hashed_password FROM users WHERE email = $1"

	err := m.DB.QueryRow(m.CTX, statement, email).Scan(&id, &hashedPassword)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrInvalidCredentials
		} else {
			return 0, err
		}
	}

	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return 0, ErrInvalidCredentials
		} else {
			return 0, err
		}
	}

	return id, nil
}

func (m *UserModel) Exists(id int) (bool, error) {
	var exists bool

	statement := "SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)"

	err := m.DB.QueryRow(m.CTX, statement, id).Scan(&exists)

	return exists, err
}
