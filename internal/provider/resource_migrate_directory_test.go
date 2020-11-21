package provider

import (
	"fmt"
	"path/filepath"
	"testing"

	helperresource "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/paultyng/terraform-provider-sql/internal/migration"
)

func TestResourceMigrateDirectory(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long test")
	}

	// currently only testing postgres
	var postgresServer *testServer
	for _, ts := range testServers {
		if ts.ServerType != "postgres" {
			continue
		}

		postgresServer = ts
		break
	}

	if postgresServer == nil {
		t.Skip("unable to find postgres server information")
	}

	for dir, c := range map[string]struct {
		split string
		table string
	}{
		"go-migrate": {"", "go_migrate_test_table"},
		"shmig":      {migration.SHMigSplit, "shmig_test_table"},
	} {
		t.Run(dir, func(t *testing.T) {

			migrationPath, err := filepath.Abs(filepath.Join("../migration/testdata", dir))
			if err != nil {
				t.Fatal(err)
			}

			url, _, err := postgresServer.URL()
			if err != nil {
				t.Fatal(err)
			}

			config := fmt.Sprintf(`
			provider "sql" {
				url = %q
			
				max_idle_conns = 0
			}
			
			resource "sql_migrate_directory" "db" {
				path              = %q
				single_file_split = %q
			}
			
			data "sql_query" "users" {
				depends_on = [sql_migrate_directory.db]
			
				query = "select * from %s"
			}
			
			output "rowcount" {
			value = length(data.sql_query.users.result)
			}
					`, url, migrationPath, c.split, c.table)

			helperresource.UnitTest(t, helperresource.TestCase{
				ProtoV5ProviderFactories: protoV5ProviderFactories,
				Steps: []helperresource.TestStep{
					{
						Config: config,
						Check: helperresource.ComposeTestCheckFunc(
							helperresource.TestCheckOutput("rowcount", "4"),
						),
					},
					// run test again to ensure no change
					{
						Config: config,
						Check: helperresource.ComposeTestCheckFunc(
							helperresource.TestCheckOutput("rowcount", "4"),
						),
					},
				},
			})
		})
	}
}
