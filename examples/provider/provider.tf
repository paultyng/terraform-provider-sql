# connect to Microsoft SQL Server
provider "sql" {
  url   = "sqlserver://sa:password@localhost:1433"
}

# connect to PostgreSQL
provider "sql" {
  url   = "postgres://postgres:password@localhost:5432/mydatabase?sslmode=disable"
}

# connect to CockroachDB
provider "sql" {
  # use the postgres driver for CockroachDB
  url = "postgres://root@localhost:26257/events?sslmode=disable"
}

# connect to a MySQL server
provider "sql" {
  url   = "mysql://root:password@tcp(localhost:3306)/mysql"
}