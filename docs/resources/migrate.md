---
page_title: "sql_migrate Resource - terraform-provider-sql"
subcategory: ""
description: |-
  
---

# Resource `sql_migrate`



## Example Usage

```terraform
resource "sql_migrate" "db" {
  migration {
    up = <<SQL
CREATE TABLE users (
	user_id integer unique,
	name    varchar(40),
	email   varchar(40)
);
SQL

    down = "DROP TABLE IF EXISTS users;"
  }

  migration {
    up   = "INSERT INTO users VALUES (1, 'Paul Tyng', 'paul@example.com');"
    down = "DELETE FROM users WHERE user_id = 1;"
  }
}

data "sql_query" "users" {
  # run this query after the migration
  depends_on = [sql_migrate.db]

  query = "select * from users"
}

output "rowcount" {
  value = length(data.sql_query.users.result)
}
```

## Schema

### Optional

- **migration** (Block List) (see [below for nested schema](#nestedblock--migration))

### Read-only

- **complete_migrations** (List of Object, Read-only) The completed migrations that have been run against your database. This list is used as storage to migrate down or as a trigger for downstream dependencies. (see [below for nested schema](#nestedatt--complete_migrations))
- **id** (String, Read-only, Deprecated) This attribute is only present for some compatibility issues and should not be used. It will be removed in a future version.

<a id="nestedblock--migration"></a>
### Nested Schema for `migration`

Required:

- **down** (String, Required) The query to run when undoing this migration.
- **id** (String, Required) Identifier can be any string to help identifying the migration in the source.
- **up** (String, Required) The query to run when applying this migration.


<a id="nestedatt--complete_migrations"></a>
### Nested Schema for `complete_migrations`

- **down** (String)
- **id** (String)
- **up** (String)


