package storage

import (
	"database/sql"
	"fmt"
	_ "github.com/jackc/pgx/v4/stdlib"
)

type Postgres struct {
	conn *sql.DB
}

func New(dsn string) (Postgres, error) {
	conn, err := sql.Open("pgx", dsn)
	if err != nil {
		return Postgres{}, fmt.Errorf("failed to open database connection: %w", err)
	}

	err = conn.Ping()
	if err != nil {
		return Postgres{}, fmt.Errorf("failed to ping database: %w", err)
	}

	err = migrateSchema(conn)
	if err != nil {
		return Postgres{}, fmt.Errorf("failed to execute schema migrations: %w", err)
	}

	return Postgres{conn: conn}, err
}

func (p Postgres) Close() {
	_ = p.conn.Close()
}
