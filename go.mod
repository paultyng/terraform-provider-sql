module github.com/paultyng/terraform-provider-sql

go 1.16

require (
	github.com/denisenkom/go-mssqldb v0.12.2
	github.com/go-sql-driver/mysql v1.6.0
	github.com/google/go-cmp v0.5.8
	github.com/hashicorp/go-argmapper v0.2.4
	github.com/hashicorp/go-plugin v1.4.4
	github.com/hashicorp/terraform-plugin-docs v0.13.0
	github.com/hashicorp/terraform-plugin-go v0.2.1
	github.com/hashicorp/terraform-plugin-sdk/v2 v2.5.0
	github.com/jackc/pgx/v4 v4.17.2
	github.com/ory/dockertest/v3 v3.9.1
)

// replace github.com/hashicorp/terraform-plugin-go => ../../hashicorp/terraform-plugin-go
// replace github.com/hashicorp/go-argmapper => ../../hashicorp/go-argmapper
