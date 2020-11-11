---
page_title: "sql Provider"
subcategory: ""
description: |-
  
---

# sql Provider



## Example Usage

```terraform
# connect to Microsoft SQL Server
provider "sql" {
  alias = "mssql"
  url   = "sqlserver://sa:password@localhost:1433"
}

# connect to PostgreSQL
provider "sql" {
  alias = "postgres"
  url   = "postgres://postgres:password@localhost:5432/mydatabase?sslmode=disable"
}

# connect to CockroachDB
provider "sql" {
  alias = "cockroach"
  # use the postgres driver for CockroachDB
  url = "postgres://root@localhost:26257/events?sslmode=disable"
}

# connect to a MySQL server
provider "sql" {
  alias = "mysql"
  url   = "mysql://root:password@tcp(localhost:3306)/mysql"
}
```

## Schema

### Optional

- **max_idle_conns** (Number, Optional) Sets the maximum number of connections in the idle connection pool. Default is `2`. See Go's documentation on [DB.SetMaxIdleConns](https://golang.org/pkg/database/sql/#DB.SetMaxIdleConns).
- **max_open_conns** (Number, Optional) Sets the maximum number of open connections to the database. Default is `0` (unlimited). See Go's documentation on [DB.SetMaxOpenConns](https://golang.org/pkg/database/sql/#DB.SetMaxOpenConns).
- **url** (String, Optional) Database connection strings are specified via URLs. The URL format is driver dependent but generally has the form: `dbdriver://username:password@host:port/dbname?param1=true&param2=false`. You can optionally set the `SQL_URL` environment variable instead.
