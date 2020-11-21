resource "sql_migrate_directory" "db" {
  # directory of *.up.sql and *.down.sql files
  path = "${path.module}/migrations"
}

data "sql_query" "users" {
  # run this query after the migration
  depends_on = [sql_migrate_directory.db]

  query = "select * from users"
}

output "rowcount" {
  value = length(data.sql_query.users.result)
}
