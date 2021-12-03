package storage

import (
	"database/sql"
	"fmt"
	"github.com/lopezator/migrator"
	"github.com/sirupsen/logrus"
)

func migrateSchema(db *sql.DB) error {
	m, err := migrator.New(
		migrator.WithLogger(migrator.LoggerFunc(func(s string, i ...interface{}) {
			logrus.Infof(s, i)
		})),
		migrator.Migrations(
			migrations...,
		),
	)
	if err != nil {
		return fmt.Errorf("failed to bootstrap database schema migration: %w", err)
	}

	err = m.Migrate(db)
	if err != nil {
		return fmt.Errorf("failed to execute database schema migration: %w", err)
	}

	return nil
}

var migrations = []interface{}{
	&migrator.MigrationNoTx{
		Name: "Setup note-lists, note and note_permissions",
		Func: func(db *sql.DB) error {
			_, err := db.Exec(`CREATE TABLE IF NOT EXISTS note_lists (
					id 			SERIAL,
                    name 		TEXT,
                    description TEXT,
            		PRIMARY KEY(id)
			);
			
			CREATE TABLE IF NOT EXISTS notes (
					id 			SERIAL,
                    name 		TEXT,
                    description TEXT,
                    list_id 	SERIAL,
            		PRIMARY KEY(id),
   					CONSTRAINT fk_list FOREIGN KEY(list_id) REFERENCES note_lists(id)
			);

			CREATE TABLE IF NOT EXISTS note_permissions (
					list_id SERIAL,
                    favored TEXT,
            		PRIMARY KEY(list_id, favored),
   					CONSTRAINT fk_list FOREIGN KEY(list_id) REFERENCES note_lists(id)
			);
			`)
			return err
		},
	},
}
