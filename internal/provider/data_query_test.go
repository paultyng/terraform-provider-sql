package provider

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestQuery_driverTypes(t *testing.T) {
	const testColName = "testcol"

	if testing.Short() {
		t.Skip("skipping long test")
	}

	for k, url := range testURLs {
		t.Run(k, func(t *testing.T) {
			scheme, err := schemeFromURL(url)
			if err != nil {
				t.Fatalf("unable to determine driver scheme for %s: %s", k, err)
			}
			var literals map[string]struct {
				sql      string
				expected string
			}
			// TODO: check output types for expected HCL type, not sure how to to do this
			switch scheme {
			case "mysql":
				literals = map[string]struct {
					sql      string
					expected string
				}{
					// https://dev.mysql.com/doc/refman/8.0/en/cast-functions.html#function_convert

					// TODO: "binary"
					"char":     {"cast('foo' as char)", "foo"},
					"date":     {"cast('2020-11-16' as date)", "2020-11-16T00:00:00Z"},
					"datetime": {"cast('2020-11-16 19:00:01' as datetime)", "2020-11-16T19:00:01Z"},
					"decimal":  {"cast(1.2 as decimal(4,3))", ""},
					"double":   {"cast(1.2 as double)", "1.2"},
					"float":    {"cast(.125 as float(5))", "0.125"},
					// TODO: parse to HCL types
					"json":     {"JSON_TYPE('[1, 2, 3]')", ""},
					"nchar":    {"cast('foo' as nchar)", "foo"},
					"real":     {"cast(.125 as real)", "0.125"},
					"signed":   {"cast(-7 as signed)", "-7"},
					"time":     {"cast('04:05:06' as time)", "04:05:06"},
					"unsigned": {"cast(1 as unsigned)", "1"},
					"year":     {"cast(2020 as year)", "2020"},
				}
			case "postgres":
				literals = map[string]struct {
					sql      string
					expected string
				}{
					"bigint":                   {"cast(1 as bigint)", "1"},
					"bit":                      {"cast(B'1001' as bit (4))", "1001"},
					"bit varying":              {"cast(B'1001' as bit varying (4))", "1001"},
					"bool":                     {"cast(true as bool)", "true"},
					"character":                {"cast('aaa' as character (3))", "aaa"},
					"character varying":        {"cast('abc def' as character varying)", "abc def"},
					"cidr":                     {"cast('192.168.1.0/24' as cidr)", "192.168.1.0/24"},
					"date":                     {"cast('1999-01-08' as date)", "1999-01-08T00:00:00Z"},
					"double precision":         {"cast(1.2 as double precision)", "1.2"},
					"inet":                     {"cast('192.168.1.1' as inet)", "192.168.1.1"},
					"integer":                  {"cast(3 as integer)", "3"},
					"macaddr":                  {"cast('08:00:2b:01:02:03' as macaddr)", "08:00:2b:01:02:03"},
					"macaddr8":                 {"cast('08:00:2b:01:02:03:04:05' as macaddr8)", "08:00:2b:01:02:03:04:05"},
					"numeric":                  {"cast(1.234 as numeric)", "1.234"},
					"real":                     {"cast(.125 as real)", "0.125"},
					"smallint":                 {"cast(12 as smallint)", "12"},
					"text":                     {"cast('foo' as text)", "foo"},
					"time":                     {"cast('04:05:06.789' as time)", "04:05:06.789"},
					"time with time zone":      {"cast('04:05:06 PST' as time with time zone)", ""},
					"timestamp":                {"cast('1999-01-08 04:05:06' as timestamp)", "1999-01-08T04:05:06Z"},
					"timestamp with time zone": {"cast('January 8 04:05:06 1999 PST' as timestamp with time zone)", "1999-01-08T07:05:06-05:00"},
					"uuid":                     {"cast('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11' as uuid)", "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11"},
					"xml":                      {`XMLPARSE (DOCUMENT '<?xml version="1.0"?><book><title>Manual</title><chapter>...</chapter></book>')`, ""},

					// TODO: money is not supported properly, just as string
					"money": {"cast('12.34' as money)", ""},

					// TODO: actually convert this to HCL types?
					"json":  {"cast('[1, 2]' as json)", ""},
					"jsonb": {"cast('[4, 5, null]' as jsonb)", ""},

					// TODO: other data types:

					// box	 	rectangular box on a plane
					// bytea	 	binary data (“byte array”)
					// TODO: "bytea": `cast('\xDEADBEEF' as bytea)`,
					// circle	 	circle on a plane
					// interval [ fields ] [ (p) ]	 	time span
					// line	 	infinite line on a plane
					// lseg	 	line segment on a plane
					// path	 	geometric path on a plane
					// pg_lsn	 	PostgreSQL Log Sequence Number
					// pg_snapshot	 	user-level transaction ID snapshot
					// point	 	geometric point on a plane
					// polygon	 	closed geometric path on a plane
					// tsquery	 	text search query
					// tsvector	 	text search document
					// txid_snapshot	 	user-level transaction ID snapshot (deprecated; see pg_snapshot)
				}

				// remove a few tests for cockroach db:
				if k == "cockroach" {
					delete(literals, "cidr")
					delete(literals, "macaddr")
					delete(literals, "macaddr8")
					delete(literals, "time with time zone")
					delete(literals, "timestamp with time zone")
					delete(literals, "xml")
				}
			case "sqlserver":
				literals = map[string]struct {
					sql      string
					expected string
				}{
					// exact numerics
					"bigint":   {"cast(12345 as bigint)", "12345"},
					"int":      {"cast(12345 as int)", "12345"},
					"smallint": {"cast(-12345 as smallint)", "-12345"},
					"tinyint":  {"cast(200 as tinyint)", "200"},

					"bit": {"cast(1 as bit)", "true"},

					// TODO: these are not yet supported properly, just as string
					"decimal":    {"cast(123.4 as decimal(9,3))", ""},
					"money":      {"cast(123.45 as money)", ""},
					"smallmoney": {"cast(-123.45 as smallmoney)", ""},

					// aproximate numerics
					"float": {"cast(.125 as float(5))", "0.125"},
					"real":  {"cast(.125 as real)", "0.125"},

					// character strings
					"char":    {"cast('a' as char(1))", "a"},
					"varchar": {"cast('abc' as varchar(10))", "abc"},
					"text":    {"cast('abcdef' as text)", "abcdef"},

					// unicode strings
					"nchar":    {"cast('a' as nchar(1))", "a"},
					"nvarchar": {"cast('abc' as nvarchar(10))", "abc"},
					"ntext":    {"cast('abcdef' as ntext)", "abcdef"},

					// other data types
					"uniqueidentifier": {"cast('0E984725-C51C-4BF4-9960-E1C80E27ABA0' as uniqueidentifier)", "0E984725-C51C-4BF4-9960-E1C80E27ABA0"},

					// TODO: other data types:

					// binary
					// varbinary
					// image
					// cursor
					// rowversion
					// hierarchyid
					// sql_variant
					// xml
					// spatial types
					// table?
				}
			}

			if len(literals) == 0 {
				t.Skipf("no literals to test defined for scheme: %q (%s)", scheme, k)
			}
			for name, lit := range literals {

				t.Run(name, func(t *testing.T) {
					// for debugging a single type...
					// if name != "money" && name != "smallmoney" && name != "decimal" {
					// 	t.Skip()
					// }

					// fix slash escaping
					col := strings.ReplaceAll(lit.sql, `\`, `\\`)
					query := fmt.Sprintf("select %s as %s", col, testColName)
					resource.UnitTest(t, resource.TestCase{
						ProtoV5ProviderFactories: map[string]func() (tfprotov5.ProviderServer, error){
							"sql": func() (tfprotov5.ProviderServer, error) {
								return New("acctest")(), nil
							},
						},
						Steps: []resource.TestStep{
							{

								Config: fmt.Sprintf(`
provider "sql" {
	url = %q

	max_idle_conns = 0
}

data "sql_query" "test" {
	query = %q
}

output "query" {
	value = data.sql_query.test.result
}
				`, url, query),
								Check: resource.ComposeTestCheckFunc(
									func(s *terraform.State) error {
										rs := s.RootModule().Resources["data.sql_query.test"]
										att := rs.Primary.Attributes["result.0."+testColName]
										if lit.expected == "" {
											t.Logf("skipping value check, but got %q", att)
										} else if lit.expected != att {
											return fmt.Errorf("expected %q, got %q", lit.expected, att)
										}
										return nil
									},
								),
							},
						},
					})
				})
			}
		})
	}
}
