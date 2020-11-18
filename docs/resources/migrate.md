---
page_title: "sql_migrate Resource - terraform-provider-sql"
subcategory: ""
description: |-
  
---

# Resource `sql_migrate`



## Example Usage

```terraform
resource "sql_migrate" "users" {
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
  depends_on = [sql_migrate.users]

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

- **id** (String, Read-only, Deprecated) The ID of this resource.

<a id="nestedblock--migration"></a>
### Nested Schema for `migration`

Required:

- **down** (String, Required)
- **up** (String, Required)

Optional:

- **id** (String, Optional) Identifier can be any string to help identifying the migration in the source.


