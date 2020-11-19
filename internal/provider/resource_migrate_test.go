package provider

import (
	"fmt"
	"testing"

	helperresource "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestResourceMigrate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long test")
	}

	for _, server := range testServers {
		t.Run(server.ServerType, func(t *testing.T) {
			url, _, err := server.URL()
			if err != nil {
				t.Fatal(err)
			}

			helperresource.UnitTest(t, helperresource.TestCase{
				ProtoV5ProviderFactories: protoV5ProviderFactories,
				Steps: []helperresource.TestStep{
					{
						Config: fmt.Sprintf(`
provider "sql" {
	url = %q

	max_idle_conns = 0
}

resource "sql_migrate" "users" {
	migration {
		id = "create table"

		up = <<SQL
CREATE TABLE users (
	user_id integer unique,
	name    varchar(40),
	email   varchar(40)
);
SQL

		down = <<SQL
DROP TABLE users;
SQL
	}
}

data "sql_query" "users" {
	depends_on = [sql_migrate.users]

	query = "select * from users"
}

output "rowcount" {
	value = length(data.sql_query.users.result)
}
				`, url),
						Check: helperresource.ComposeTestCheckFunc(
							helperresource.TestCheckOutput("rowcount", "0"),
						),
					},
					{
						Config: fmt.Sprintf(`
provider "sql" {
	url = %q

	max_idle_conns = 0
}

resource "sql_migrate" "users" {
	migration {
		id = "create table"

		up = <<SQL
CREATE TABLE users (
	user_id integer unique,
	name    varchar(40),
	email   varchar(40)
);
SQL

		down = <<SQL
DROP TABLE users;
SQL
	}

	migration {
		id   = "insert row"
		up   = "INSERT INTO users VALUES (1, 'Paul Tyng', 'paul@example.com');"
		down = "DELETE FROM users WHERE user_id = 1;"
	}
}

data "sql_query" "users" {
	depends_on = [sql_migrate.users]

	query = "select * from users"
}

output "rowcount" {
	value = length(data.sql_query.users.result)
}
				`, url),
						Check: helperresource.ComposeTestCheckFunc(
							helperresource.TestCheckOutput("rowcount", "1"),
						),
					},
				},
			})
		})
	}
}
