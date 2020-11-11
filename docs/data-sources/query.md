---
page_title: "sql_query Data Source - terraform-provider-sql"
subcategory: ""
description: |-
  The sql_query datasource allows you to execute a SQL query against the database of your choice.
---

# Data Source `sql_query`

The `sql_query` datasource allows you to execute a SQL query against the database of your choice.

## Example Usage

```terraform
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
```

## Schema

### Required

- **query** (String, Required) The query to execute. The types in this query will be reflected in the typing of the `result` attribute.

### Read-only

- **id** (String, Read-only, Deprecated) This attribute is only present for some compatibility issues and should not be used. It will be removed in a future version.
- **result** (List of Dynamic, Read-only) The result of the query. This will be a list of objects. Each object will have attributes with names that match column names and types that match column types. The exact translation of types is dependent upon the database driver.


