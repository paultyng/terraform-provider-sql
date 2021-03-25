module github.com/paultyng/terraform-provider-sql

go 1.15

require (
	github.com/Microsoft/go-winio v0.4.15 // indirect
	github.com/containerd/continuity v0.0.0-20200928162600-f2cc35102c2a // indirect
	github.com/denisenkom/go-mssqldb v0.9.0
	github.com/go-sql-driver/mysql v1.5.0
	github.com/google/go-cmp v0.5.4
	github.com/hashicorp/go-argmapper v0.0.0-20200721221215-04ae500ede3b
	github.com/hashicorp/go-plugin v1.4.0
	github.com/hashicorp/terraform-plugin-docs v0.3.1-0.20210107204619-bf524a84dc08
	github.com/hashicorp/terraform-plugin-go v0.2.1
	github.com/hashicorp/terraform-plugin-sdk/v2 v2.5.0
	github.com/jackc/pgx/v4 v4.10.1
	github.com/ory/dockertest/v3 v3.6.3
	gopkg.in/yaml.v2 v2.2.8 // indirect
)

// replace github.com/hashicorp/terraform-plugin-go => ../../hashicorp/terraform-plugin-go
