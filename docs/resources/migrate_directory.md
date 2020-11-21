---
page_title: "sql_migrate_directory Resource - terraform-provider-sql"
subcategory: ""
description: |-
  
---

# Resource `sql_migrate_directory`



## Example Usage

```terraform
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
```

## Schema

### Required

- **path** (String, Required) The path of the SQL migration files. For a path relative to the current module, use `path.module`.

### Optional

- **single_file_split** (String, Optional) Set this to a value if your migration up and down are in a single file, split on some constant string (ie. in the case of [shmig](https://github.com/mbucc/shmig) you would use `-- ==== DOWN ====`).

### Read-only

- **complete_migrations** (List of Object, Read-only) The completed migrations that have been run against your database. This list is used as storage to migrate down or as a trigger for downstream dependencies. (see [below for nested schema](#nestedatt--complete_migrations))
- **id** (String, Read-only, Deprecated) This attribute is only present for some compatibility issues and should not be used. It will be removed in a future version.

<a id="nestedatt--complete_migrations"></a>
### Nested Schema for `complete_migrations`

- **down** (String)
- **id** (String)
- **up** (String)


