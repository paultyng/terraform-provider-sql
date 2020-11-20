package migration

import (
	"context"
	"database/sql"
)

type Migration struct {
	ID   string
	Up   string
	Down string
}

func Subtract(x, y []Migration) []Migration {
	result := []Migration{}
	for _, xm := range x {
		for _, ym := range y {
			if xm.ID == ym.ID {
				goto NextMigration
			}
		}

		result = append(result, xm)

	NextMigration:
	}

	return result
}

func Up(ctx context.Context, db *sql.DB, all, applied []Migration) error {
	removedMigrations := Subtract(applied, all)
	newMigrations := Subtract(all, applied)

	err := runMigrations(ctx, false, removedMigrations, execMigration(db))
	if err != nil {
		return nil
	}

	err = runMigrations(ctx, true, newMigrations, execMigration(db))
	if err != nil {
		return err
	}

	return nil
}

func Down(ctx context.Context, db *sql.DB, all, applied []Migration) error {
	return runMigrations(ctx, false, applied, execMigration(db))
}

func execMigration(db *sql.DB) func(context.Context, Migration, string) error {
	return func(ctx context.Context, m Migration, query string) error {
		_, err := db.ExecContext(ctx, query)
		if err != nil {
			return err
		}
		return nil
	}
}

func runMigrations(ctx context.Context, up bool, migrations []Migration, run func(context.Context, Migration, string) error) error {
	// TODO: add diagnostics to track down specific migration at error
	var err error

	if up {
		for _, m := range migrations {
			err = run(ctx, m, m.Up)
			if err != nil {
				return err
			}
		}
	} else {
		for i := len(migrations) - 1; i >= 0; i-- {
			m := migrations[i]

			err = run(ctx, m, m.Down)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
