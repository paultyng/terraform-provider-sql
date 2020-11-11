data "sql_query" "test" {
  query = "select 1 as number, 'foo' as string"
}

locals {
  # The number column in this case is a Terraform "Number" type
  # so you can use it as such:
  math = 1 + data.sql_query.test.result[0].number
}

output "math" {
  value = local.math
}