module github.com/paultyng/terraform-provider-sql

go 1.15

require (
	github.com/Microsoft/go-winio v0.4.15 // indirect
	github.com/containerd/continuity v0.0.0-20200928162600-f2cc35102c2a // indirect
	github.com/denisenkom/go-mssqldb v0.9.0
	github.com/go-sql-driver/mysql v1.5.0
	github.com/google/go-cmp v0.5.3
	github.com/hashicorp/go-plugin v1.4.0
	github.com/hashicorp/terraform-plugin-docs v0.2.0
	github.com/hashicorp/terraform-plugin-go v0.1.1-0.20201117024036-b9d161518a6d
	github.com/hashicorp/terraform-plugin-sdk/v2 v2.3.0
	github.com/jackc/pgx/v4 v4.9.2
	github.com/ory/dockertest/v3 v3.6.2
	gopkg.in/yaml.v2 v2.2.8 // indirect
)

// replace github.com/hashicorp/terraform-plugin-go => ../../hashicorp/terraform-plugin-go
