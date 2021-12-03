package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/jackc/pgconn"
)

var ErrNoteNotFound = errors.New("requested note does not exists")
var ErrNoteListNotFound = errors.New("requested note-list does not exists")

type Note struct {
	ID          int32
	Name        string
	Description string
	ListID      int32
}

func (p Postgres) CreateNote(note Note, authorizations []string) (int32, error) {
	tx, err := p.conn.Begin()
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}

	err = tx.QueryRow(
		"SELECT true FROM note_permissions WHERE list_id = $1 AND favored = ANY($2::text[]);",
		note.ListID, authorizations).Err()
	if err != nil {
		_ = tx.Rollback()
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrNoteListNotFound
		}
		return 0, fmt.Errorf("failed to check permissions to insert note: %w", err)
	}

	var noteID int32
	err = tx.QueryRow(`INSERT INTO notes (name, description, list_id)
							VALUES ($1, $2, $3) RETURNING id
							`, note.Name, note.Description, note.ListID).Scan(&noteID)
	if err != nil {
		_ = tx.Rollback()
		pgErr, isPGErr := err.(*pgconn.PgError)
		if isPGErr && pgErr.ConstraintName == "fk_list" {
			return 0, ErrNoteListNotFound
		}
		return 0, fmt.Errorf("failed to insert note: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return 0, fmt.Errorf("failed to commit insert note transaction: %w", err)
	}

	return noteID, err
}

func (p Postgres) Note(noteID int32, authorizations []string) (Note, error) {
	var note Note

	err := p.conn.QueryRow(`SELECT n.id, n.list_id, n.name, n.description
							FROM notes AS n
         					FULL OUTER JOIN note_permissions AS p ON n.list_id = p.list_id
							WHERE n.id = $1 AND p.favored = ANY($2::text[]);`, noteID, authorizations).
		Scan(&note.ID, &note.ListID, &note.Name, &note.Description)
	if err != nil {
		return Note{}, fmt.Errorf("failed to scan note-query result: %w", err)
	}

	return note, err
}

func (p Postgres) UpdateNote(note Note, authorizations []string) error {
	res, err := p.conn.Exec(`UPDATE notes SET name = $2, description = $3
			WHERE 
				id = $1 AND 
			    list_id IN (
			        SELECT list_id FROM note_permissions AS p WHERE p.favored = ANY($4::text[])
				)`, note.ID, note.Name, note.Description, authorizations)
	if err != nil {
		return fmt.Errorf("failed to update note: %w", err)
	}

	affectedRow, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed calculate count of updated notes: %w", err)
	}

	if affectedRow < 1 {
		return ErrNoteNotFound
	}

	return nil
}

func (p Postgres) DeleteNote(noteID int32, authorizations []string) error {
	res, err := p.conn.Exec(`DELETE FROM notes WHERE 
				id = $1 AND 
			    list_id IN (
			        SELECT list_id FROM note_permissions AS p WHERE p.favored = ANY($2::text[])
				)`, noteID, authorizations)
	if err != nil {
		return fmt.Errorf("failed to update note: %w", err)
	}

	affectedRow, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed calculate count of updated notes: %w", err)
	}

	if affectedRow < 1 {
		return ErrNoteNotFound
	}
	return nil
}
