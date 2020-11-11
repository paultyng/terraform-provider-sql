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