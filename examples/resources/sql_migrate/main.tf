terraform {
  required_providers {
    sql = {
      source = "paultyng/sql"
    }
  }
}

provider "sql" {
  url = "postgres://postgres:tf@localhost:5432/tftest?sslmode=disable"
}
