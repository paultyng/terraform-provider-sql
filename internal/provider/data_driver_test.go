package provider

import (
	"fmt"
	"testing"

	helperresource "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestDataDriver(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long test")
	}

	for _, server := range testServers {
		// TODO: check nulls for all these
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

data "sql_driver" "test" {
}
		`, url),
						Check: helperresource.ComposeTestCheckFunc(
							func(s *terraform.State) error {
								rs := s.RootModule().Resources["data.sql_driver.test"]
								att := rs.Primary.Attributes["name"]
								if att != server.ExpectedDriver {
									return fmt.Errorf("expected %q, got %q", server.ExpectedDriver, att)
								}
								return nil
							},
						),
					},
				},
			})
		})
	}
}
