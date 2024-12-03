package models

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Define a Snippet type to hold the data for an individual snippet. Notice how the fields of the struct correspond to the fields in the database
type Snippet struct {
	ID      int
	Title   string
	Content string
	Created time.Time
	Expires time.Time
}

// The snippet model will be responsible for interacting with the DB, like inserting, updating, deleting, etc
type SnippetModel struct {
	DB  *pgxpool.Pool
	CTX context.Context
}

func (m *SnippetModel) Insert(title, content string, expires time.Time) (int, error) {
	// using blockquotes for readability, so we can break the lines
	// we use RETURNING to get the newly added ID
	statement := `INSERT INTO snippets (title, content, created, expires)
  VALUES ($1, $2, CURRENT_TIMESTAMP, $3) RETURNING id`

	var newId int

	// notice that instead of .Exec() we use .QueryRow() because we are using RETURNING
	err := m.DB.QueryRow(m.CTX, statement, title, content, expires).Scan(&newId)
	if err != nil {
		return 0, err
	}

	return newId, nil
}

func (m *SnippetModel) Get(id int) (Snippet, error) {
	statement := `SELECT id, title, content, created, expires FROM snippets WHERE expires > CURRENT_TIMESTAMP AND id = $1`

	rows, _ := m.DB.Query(m.CTX, statement, id)
	snippet, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[Snippet])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Why we’re returning the ErrNoRecord error from our SnippetModel.Get() method, instead of pgx.ErrNoRows directly? The reason is to help encapsulate the model completely, so that our handlers aren’t concerned with the underlying datastore or reliant on datastore-specific errors (like sql.ErrNoRows) for its behavior.
			return Snippet{}, ErrNoRecord
		}

		return Snippet{}, err
	}

	return snippet, nil
}

func (m *SnippetModel) Latest() ([]Snippet, error) {
	statement := `SELECT id, title, content, created, expires from snippets
  WHERE expires > CURRENT_TIMESTAMP ORDER BY id DESC limit 10`

	rows, _ := m.DB.Query(m.CTX, statement)
	snippets, err := pgx.CollectRows(rows, pgx.RowToStructByName[Snippet])
	if err != nil {
		return nil, err
	}

	return snippets, nil
}
