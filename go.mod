module github.com/paultyng/terraform-provider-sql

go 1.15

require (
	github.com/Azure/go-ansiterm v0.0.0-20170929234023-d6e3b3328b78 // indirect
	github.com/Microsoft/go-winio v0.4.15 // indirect
	github.com/Nvveen/Gotty v0.0.0-20120604004816-cd527374f1e5 // indirect
	github.com/cenkalti/backoff v2.2.1+incompatible // indirect
	github.com/containerd/continuity v0.0.0-20200928162600-f2cc35102c2a // indirect
	github.com/denisenkom/go-mssqldb v0.9.0
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-units v0.4.0 // indirect
	github.com/go-sql-driver/mysql v1.5.0
	github.com/gotestyourself/gotestyourself v2.2.0+incompatible // indirect
	github.com/hashicorp/go-plugin v1.3.0
	github.com/hashicorp/terraform-plugin-docs v0.2.0
	github.com/hashicorp/terraform-plugin-go v0.1.1-0.20201117024036-b9d161518a6d
	github.com/hashicorp/terraform-plugin-sdk/v2 v2.2.0
	github.com/jackc/pgx/v4 v4.9.2
	github.com/opencontainers/image-spec v1.0.1 // indirect
	github.com/opencontainers/runc v0.1.1 // indirect
	github.com/ory/dockertest v3.3.5+incompatible
	golang.org/x/sys v0.0.0-20200826173525-f9321e4c35a6 // indirect
	gopkg.in/yaml.v2 v2.2.8 // indirect
	gotest.tools v2.2.0+incompatible // indirect
)

// see https://github.com/ory/dockertest/issues/204
replace golang.org/x/sys => golang.org/x/sys v0.0.0-20190830141801-acfa387b8d69

// replace github.com/hashicorp/terraform-plugin-go => ../../hashicorp/terraform-plugin-go
