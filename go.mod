module github.com/paultyng/terraform-provider-sql

go 1.16

require (
	github.com/containerd/continuity v0.0.0-20200928162600-f2cc35102c2a // indirect
	github.com/denisenkom/go-mssqldb v0.10.0
	github.com/go-sql-driver/mysql v1.6.0
	github.com/google/go-cmp v0.5.6
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-argmapper v0.1.1
	github.com/hashicorp/go-plugin v1.4.1
	github.com/hashicorp/terraform-plugin-docs v0.5.0
	github.com/hashicorp/terraform-plugin-go v0.2.1
	github.com/hashicorp/terraform-plugin-sdk/v2 v2.5.0
	github.com/jackc/pgx/v4 v4.11.0
	github.com/ory/dockertest/v3 v3.6.3
)

// replace github.com/hashicorp/terraform-plugin-go => ../../hashicorp/terraform-plugin-go
// replace github.com/hashicorp/go-argmapper => ../../hashicorp/go-argmapper
