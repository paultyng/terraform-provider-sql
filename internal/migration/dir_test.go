package migration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestReadDir_dirDoesNotExist(t *testing.T) {
	_, err := ReadDir("this-dir-does-not-exist", nil)
	if err == nil {
		t.Fatalf("expected error but got none")
	}

	if !os.IsNotExist(err) {
		t.Fatalf("expected path not exist but got %T %s", err, err)
	}
}

func TestReadDir(t *testing.T) {
	for dir, c := range map[string]struct {
		Options  *Options
		Expected []Migration
	}{
		"go-migrate": {
			nil,
			[]Migration{
				{
					ID: "1085649617_create_users_table",
					Up: strings.TrimSpace(`
CREATE TABLE go_migrate_test_table (
  user_id integer unique,
  name    varchar(40),
  email   varchar(40)
);				  
`),
					Down: "DROP TABLE go_migrate_test_table;",
				},
				{
					ID:   "1185749658_add_city_to_users",
					Up:   "ALTER TABLE go_migrate_test_table ADD COLUMN city varchar(100);",
					Down: "ALTER TABLE go_migrate_test_table DROP COLUMN city;",
				},
				{
					ID: "1485648520-testdata",
					Up: strings.TrimSpace(`
INSERT INTO go_migrate_test_table (user_id, name, email) VALUES (1, 'Foo Bar', 'foo@example.com');
INSERT INTO go_migrate_test_table (user_id, name, email) VALUES (2, 'Bar Baz', 'bar@example.com');
INSERT INTO go_migrate_test_table (user_id, name, email) VALUES (3, 'Baz Qux', 'baz@example.com');
INSERT INTO go_migrate_test_table (user_id, name, email) VALUES (4, 'Paul Tyng', 'paul@example.com');
`),
					Down: "DELETE FROM go_migrate_test_table;",
				},
			},
		},
		"shmig": {
			&Options{
				SingleFileSplit:   SHMigSplit,
				StripLineComments: true,
			},
			[]Migration{
				{
					ID: "1485643154-create_table",
					Up: strings.TrimSpace(`
BEGIN;

CREATE TABLE
	shmig_test_table
(
	id	   integer
	, code     varchar(200)
	, name     varchar(200)
);

COMMIT;
`),
					Down: strings.TrimSpace(`
BEGIN;

DROP TABLE
	shmig_test_table;

COMMIT;
`),
				},
				{
					ID: "1485648520-testdata",
					Up: strings.TrimSpace(`
BEGIN;

INSERT INTO shmig_test_table (code, name) VALUES ('QB' , 'Tom Brady');
INSERT INTO shmig_test_table (code, name) VALUES ('TE' , 'Ben Coates');
INSERT INTO shmig_test_table (code, name) VALUES ('CB' , 'Raymond Clayborn');
INSERT INTO shmig_test_table (code, name) VALUES ('G' ,  'John (Hog) Hannah');

COMMIT;
`),
					Down: strings.TrimSpace(`
BEGIN;

DELETE FROM shmig_test_table;

COMMIT;
`),
				},
			},
		},
	} {
		t.Run(dir, func(t *testing.T) {
			actual, err := ReadDir(filepath.Join("testdata", dir), c.Options)
			if err != nil {
				t.Fatalf("error from Read %T: %s", err, err)
			}

			if !cmp.Equal(c.Expected, actual, crlfComparer) {
				t.Fatalf("migrations do not match:\n%s", cmp.Diff(c.Expected, actual, crlfComparer))
			}
		})
	}
}

var crlfComparer = cmp.Comparer(func(x, y string) bool {
	return strings.ReplaceAll(x, "\r\n", "\n") == strings.ReplaceAll(y, "\r\n", "\n")
})
